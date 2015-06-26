#!/bin/bash

# audit system data and save it in dcman

URL={{.URL}}
[[ $DEBUG ]] && echo $URL

unset CHASSIS

list() {
        echo Hostname=$HOSTNAME

        IPS=($(ifconfig -a | awk '/inet/ {print $2}' | grep -v "192.168" | sed -e 's/addr://g' -e '/^127.*/d'))
        echo IPs=$(IFS="," ; echo "${IPS[*]}")

        MACS=($(awk '/Permanent HW addr/ {print $4}' /proc/net/bonding/bond0))
        NICS=($(awk '/Slave Interface/   {print $3}' /proc/net/bonding/bond0 | awk '{print toupper(substr($0,0,1)) substr($0,2)}'))
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
        done < <(dmidecode)

        ipmitool lan print 1 2>/dev/null | awk '/IP Address  |MAC Address/ {print "IPMI_" $1 "=" tolower($4)}'

        echo CPU=$(grep 'model name'  /proc/cpuinfo | sed -e 's/^.*: //g' -e 's/.*CPU //g' | sort -u)
        echo Mem=$(echo "$(awk '/MemTotal:/ {print $2}' /proc/meminfo) / (1024 * 1000)" | bc)
}

[[ $1 == "-l" || $DEBUG ]] && list && exit

IFS=$'\n'
ARGS=($(list))
for (( I=1 ; I < ${#ARGS[@]} ; I++ ))
do
   ARGS[$I]="-d${ARGS[$I]}"
done
curl --data-urlencode ${ARGS[*]} $URL

