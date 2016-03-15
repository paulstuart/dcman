#!/bin/bash

# print drive size, location (group,shelf,arm), s/n

IP=$(ip route get 1 | awk '{print $NF;exit}')

fail() {
    cat >&2 <<EOF
{
  "Hostname": "${HOSTNAME}",
  "IP": "${IP}",
  "Error": "$*"
}
EOF
    exit 1
}

#fail blah

export PATH=$PATH:/opt/MegaRAID/MegaCli:.

MEGACLI=$(which MegaCli64 2> /dev/null) || fail "MegaCli not installed"

data() {
sudo $MEGACLI -pdlist -aALL | awk '
/^Raw Size:/ {printf "\t%s%s\t", $3, $4};
/Span:/      {printf gensub(/[^0-9,]*/,"","g") };
/^Inquiry /  {print $3, $4, $5};
'
}

disks() {
    while read LOCATION SIZE MFGR PN SN
    do
        # deal with broken LSI reporting
        [[ $MANY ]] && echo "    ,"
        if [[ $SN =~ \. ]]; then
            SN=$MFGR
            case ${PN:0:2} in
                WD) MFGR="Western Digital" ;;
                *) unset MFGR ;;
            esac
        fi 

        cat <<EOF
    {
        "Size": "$SIZE",
        "Location": "$LOCATION",
        "Manufacturer": "$MFGR",
        "PartNumber": "$PN",
        "SerialNumber": "$SN"
    }
EOF
        MANY=true
    done < <(data)
}

json() {
    cat <<-EOF
{
    "Hostname": "${HOSTNAME}",
    "IP": "${IP}",
    "Disks": [
	$(disks)
    ]
}
	EOF
}

json 

