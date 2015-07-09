#!/bin/bash

# audit system data and save it in dcman

URL={{.URL}}
[[ $DEBUG ]] && echo $URL

unset CHASSIS

ifcfg() {
    local DOMAIN=$1
    local IFCFG=$2
    IP=$(virt-cat -d $DOMAIN /etc/sysconfig/network-scripts/$IFCFG | awk -F= '/^IPADDR/ {print $2}' | tr -d '"')
    [[ $IP ]] && echo $DOMAIN $IP
}

show() {
    local DOMAIN=$1
    IFGS=($(virt-ls -d $DOMAIN /etc/sysconfig/network-scripts/ | grep "^ifcfg" | grep -v ifcfg-lo))
    for IFCFG in ${IFGS[@]}
    do
        ifcfg $DOMAIN $IFCFG &
    done
    wait
}

kvms() {
    # make sure we have required binaries
    [[ $(which virsh 2> /dev/null) ]] || return
    [[ $(which virt-ls 2> /dev/null) ]] || return

    while read DOMAIN
    do
        show $DOMAIN &
    done < <(virsh list | tail -n +3 | sed -e '/^\s*$/d' | awk '{print $2}')
    wait
}

list() {
        echo Hostname=$HOSTNAME

        IPS=($(ifconfig -a | awk '/inet/ {print $2}' | grep -v "192.168" | sed -e 's/addr://g' -e '/^127.*/d'))
        echo IPs=$(IFS="," ; echo "${IPS[*]}")

        if [[ -f /proc/net/bonding/bond0 ]]; then
            MACS=($(awk '/Permanent HW addr/ {print $4}' /proc/net/bonding/bond0))
            NICS=($(awk '/Slave Interface/   {print $3}' /proc/net/bonding/bond0 | awk '{print toupper(substr($0,0,1)) substr($0,2)}'))
        else
            MACS=($(ifconfig -a | awk '/^eth[01].*HWaddr/ {print $5}'))
            NICS=($(ifconfig -a | awk '/^eth[01].*HWaddr/ {print toupper(substr($1,0,1)) substr($1,2)}'))
        fi
        for (( I=0 ; $I < ${#MACS[@]} ; I++))
        do
                echo ${NICS[$I]}=${MACS[$I]}
        done

        RE_ASSET="Asset Tag: ([a-zA-Z0-9-]*)"
        RE_SERIAL="Serial Number: ([a-zA-Z0-9-]*)"

        while read LINE
        do
          [[ $LINE =~ "Chassis Information" ]] && CHASSIS=true && continue
          [[ $CHASSIS ]] || continue
          [[ $LINE =~ $RE_ASSET ]] && ASSET=${BASH_REMATCH[1]} && [[ -n $ASSET && $ASSET != "To" ]] && echo Asset=$ASSET
          [[ $LINE =~ $RE_SERIAL ]] && SN=${BASH_REMATCH[1]} && [[ -n $SN && $SN != "To" ]] && echo SN=$SN  # filter out "To Be Filled By O.E.M."
          [[ $LINE ]] || break
        done < <(dmidecode 2>/dev/null)

        ipmitool lan print 1 2>/dev/null | awk '/IP Address  |MAC Address/ {print "IPMI_" $1 "=" tolower($4)}'

        echo CPU=$(grep 'model name'  /proc/cpuinfo | sed -e 's/^.*: //g' -e 's/.*CPU //g' | sort -u)
        echo Mem=$(echo "$(awk '/MemTotal:/ {print $2}' /proc/meminfo) / (1024 * 1000)" | bc)
        echo Release=$(cat /etc/*release | head -1)
        echo Kernel=$(uname -r)
}

[[ $1 == "-l" || $DEBUG ]] && list && exit

IFS=$'\n'
ARGS=($(list))
for (( I=1 ; I < ${#ARGS[@]} ; I++ ))
do
   ARGS[$I]="-d${ARGS[$I]}"
done
curl --data-urlencode ${ARGS[*]} -d VMs="$(kvms)" $URL && echo "ok" || echo "failed"



