'use strict';

var pingURL = "http://10.100.182.16:8080/dcman/api/pings?debug=true";

var deviceURL       = '/dcman/api/device/view/'
var deviceListURL   = '/dcman/api/device/ips/'
var deviceNetworkURL = '/dcman/api/device/network/'
var deviceTypesURL  = '/dcman/api/device/type/'
var ipURL           = '/dcman/api/network/ip/'
var iptypesURL      = '/dcman/api/network/ip/type/'
var ifaceURL        = '/dcman/api/interface/'
var ifaceViewURL    = '/dcman/api/interface/view/'
var mfgrURL         = '/dcman/api/mfgr/'
var inURL           = "/dcman/api/inventory/";
var vmURL           = "/dcman/api/vm/";
var vmViewURL       = "/dcman/api/vm/view/";
var partTypesURL    = "/dcman/api/part/type/";
var partURL         = "/dcman/api/part/view/";
var rackURL         = "/dcman/api/rack/view/";
var rmaURL          = "/dcman/api/rma/";
var rmaviewURL      = "/dcman/api/rma/view/";
var tagURL          = "/dcman/api/tag/";
var sitesURL        = "/dcman/api/site/" ; 
var networkURL      = "/dcman/api/network/ip/used/";
var userURL         = "/dcman/api/user/" ; 
var vlanURL         = "/dcman/api/vlan/view/" ; 
var vendorURL       = "/dcman/api/vendor/" ; 

var userInfo = {};

var mySTI = 1;

var rackData = {
    STI: 0,
    list: [],
}

function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

var getIt = function(geturl, what) {
    return function(id, query) {
        var url = geturl;
        if (query) {
            url += query
            if (id) {
                url += id
            }
        } else if (id && id > 0) {
            url += id
        }
        return get(url).then(function(result) {
            console.log('fetched:', what)
            return result;
        })
        .catch(function(x) {
            console.log('fetch failed for:', what, 'because:', x);
        });
    }
}

var commonListMIX = {
    computed: {
        sitename: function() {
            for (var i=0; i<this.sites.length; i++) {
                if (this.sites[i].STI == this.STI) {
                    return this.sites[i].Name
                }
            }
            return "ALL"
        },
    },
}

// TODO: these should be generated from a factory function
function getSiteLIST(all) {
    return get(sitesURL).then(function(result) {
        console.log('sitelist fetched:', result.length);
        if (all) {
            result.unshift({STI:0, Name:'All Sites'})
        }
        return result;
    })
    .catch(function(x) {
      console.log('Could not load sitelist: ', x);
    });
}

var getVendorList = getIt(vendorURL, 'vendors');
var getPart = getIt(partURL, 'parts');
var getPartTypes = getIt(partTypesURL, 'part typess');
var getInventory = getIt(inURL, 'inventory');
var getDeviceTypes = getIt(deviceTypesURL, 'device typess');
var getIPTypes = getIt(iptypesURL, 'ip types');
var getDeviceLIST = getIt(deviceListURL, 'device list');
var getTagList = getIt(tagURL, 'tags');
var getDevice = getIt(deviceURL, 'device');
var getMfgr = getIt(mfgrURL, 'mfgr');
var getVM = getIt(vmViewURL, 'vm');
var getRack = getIt(rackURL, 'racks')
var getRMA = getIt(rmaviewURL, 'rma')
var getVLAN = getIt(vlanURL, 'vlan')



var getInterfaces = function(device) {
    var url = ifaceViewURL + '?DID=' + device.DID; 
    return get(url).then(function(iface) {
        if (! iface) {
            device.ips = []
            device.interfaces = []
            return device
        }
        var ips = []
        var ports = {}
        for (var i=0; i<iface.length; i++) {
            var ip = iface[i]
            if (! (ip.IFD in ports)) {
                ports[ip.IFD] = ip
            }
            if (ip.IP) ips.push(ip)
        }
        var good = [];
        for (var ifd in ports) {
            good.push(ports[ifd])
        }
        device.ips = ips
        device.interfaces = good
        return device
   })
}

var deviceRacks = function(device) {
    var url = rackURL + '?STI=' + device.STI; 
    return get(url).then(function(racks) {
        device.racks = racks
        return device
    })
}


// device info with associated interface / IPs
var completeDevice = function(DID) {
   return getDevice(DID).then(getInterfaces).then(deviceRacks); 
}


function remember() {
    var cookies = document.cookie.split("; ");
    for (var i=0; i < cookies.length; i++) {
        var tuple = cookies[i].split('=')
        if (tuple[0] === 'X-API-KEY') {
            // all changeable actions require this key
            window.user_apikey = tuple[1]; 
            break
        } 
    }
}

remember();

var siteMIX = {
    route: { 
          data: function (transition) {
            //var userId = transition.to.params.userId
            return Promise.all([
                getSiteLIST(), 
           ]).then(function (data) {
              return {
                sites: data[0],
              }
            })
          }
    },
}

var validIP = function(ip) {
    if (! ip || ip.length === 0) return false;
    var octs = ip.split('.')
    if (octs.length != 4) return false
    for (var i=0; i<4; i++) {
        if (octs[i].length == 0) return false
        var val=parseInt(octs[i]);
        if (val != octs[i]) return false
        if (val < 0 || val > 255) return false
    }
    return true 
}

var ip32 = function(ip) {
    var sum = 0;
    var octs = ip.split('.');
    if (octs.length != 4) return 0
    for (var i=0; i<4; i++) {
        sum = sum << 8
        var val=parseInt(octs[i])
        if (val < 0 || val > 255) return 0
        sum += val
    }
    return sum 
}

var ipv4 = function(ip) {
    var d = ip & 255;
    var c = (ip >> 8)  & 255
    var b = (ip >> 16) & 255
    var a = (ip >> 24) & 255
    return a + "." + b + "." + c + "." + d
}

var saveMe = function(url, data, id, fn) {
    if (id && id > 0) {
        postIt(url + id + "?debug=true", data, fn, 'PATCH')
    } else {
        postIt(this.dataURL + id + "?debug=true", data, fn)
    }
}

// Create the view-model
var pagedCommon = {
    data: function() {
        return {
            rows: [],
            columns: [],
            searchQuery: '',
            startRow: 0,
            pagerows: 25,
            sizes: [10, 25, 50, 100, 'all'],
        }
    },
    computed: {
        rowsPerPage: function() {
            if (this.pagerows == 'all') return null;
            return parseInt(this.pagerows);
        },
    },
    methods: {
        resetStartRow: function() {
            this.startRow = 0;
        },
    },
}

// common stuff for edits 
var editVue = {
    methods: {
        saveSelf: function() {
            var data = this.myself()
            var id = this.myID()
            if (id > 0) {
                postIt(this.dataURL + id + "?debug=true", data, this.showList, 'PATCH')
            } else {
                postIt(this.dataURL + id + "?debug=true", data, this.showList)
            }
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            postIt(this.dataURL + this.myID(), null, this.showList, 'DELETE')
        },
        showList: function(ev) {
            router.go(this.listURL)
        },
    }
}

var tableTmpl = {
    props: ['columns', 'rows'],
    data: function () {
        return {
            sortKey: '',
            sortOrders: []
        }
    },
    methods: {
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
    },
    watch: {
        'columns': function() {
            var sortOrders = {}
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
            this.sortOrders = sortOrders
        }
    },
}



var noLinks = {
    methods: {
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
        linkable: function(key) {
            return false
        },
        linkpath: function(entry, key) {
        }
    },
}

var foundVue = {
    props: ['columns', 'rows'],
    data: function () {
        return {
            sortKey: '',
            sortOrders: []
        }
    },
    methods: {
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            if (entry.Kind.toLowerCase() === 'vm') {
                return '/vm/edit/' + entry['ID']
            }
            return '/device/edit/' + entry['ID']
        }
    },
    watch: {
        'columns': function() {
            var sortOrders = {}
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
        }
    },
    events: {
        'found-these': function(funny) {
            console.log("found-these found:", funny.length)
            this.rows = funny
        }
    }
}


var fList = Vue.component('found-list', {
    template: '#tmpl-base-table',
    mixins: [foundVue],
})

// remove leading/trailing spaces, non-ascii
var cleanText = function(text) {
    text = text.replace(/^[^A-Za-z0-9:\.\-]*/g, '')
    text = text.replace(/[^A-Za-z0-9:\.\-]*$/g, '')
    text = text.replace(/[^A-Za-z 0-9:\.\-]*/g, '')
    return text
}

var searchVue = {
    data: function () {
        return {
            columns: ['Kind', 'Name'],
            searchText: '',
            found: [],
        }
    },
    route: { 
        data: function (transition) {
            this.search(this.$route.params.searchText)
        }
    },
    methods: {
        search: function(what) {
            console.log("DO SEARCH:",what)
            this.searchText = what
            var self = this;
            if (what.length === 0) { 
                return
            }
            var searchURL = "/dcman/api/search/";
            var url = searchURL + what;
            fetchData(url, function(data) {
                console.log('we are searching for:', what)
                if (data) {
                    console.log('search matched:', data.length)
                    if (data.length == 1) {
                        console.log('what:', what)
                        if (data[0].Kind.toLowerCase() === 'vm') {
                            router.go('/vm/edit/' + data[0].ID)
                            self.$dispatch('vm-found', 'yes please')
                            return
                        }
                        router.go('/device/edit/' + data[0].ID)
                    } else {
                        self.found = data
                    }
                } else {
                    self.found = []
                }
            })
        }
    },
    events: {
        'search-again': function(text) {
            this.searchText = text;
            this.search(text)
        },
    },
}


var sList = Vue.component('search-for', {
    template: '#tmpl-search-for',
    mixins: [searchVue],
})


Vue.component('my-nav', {
    template: '#tmpl-main-menu',
    props: ['app', 'msg'],
    data: function() {
       return {
           searchText: '',
            /*
           found: [],
           columns: ['Kind', 'Name'],
            */
       }
    },
    created: function() {
        this.userinfo()
    },
    methods: {
        'doSearch': function(ev) {
            var text = cleanText(this.searchText);
            //if (this.searchText.length > 0) {
            if (text.length > 0) {
               // console.log('search for:',this.searchText)
                console.log('initiate search for:',text)
                if (this.$route.name == 'search') {
                    // already on search page
                    //this.$dispatch('search-again', this.searchText)
                    this.$dispatch('search-again', text)
                    return
                }
                //router.go({name: 'search', params: { searchText: this.searchText }})
                router.go({name: 'search', params: { searchText: text }})
            }
        },
        'userinfo': function() {
            var cookies = document.cookie.split("; ");
            for (var i=0; i < cookies.length; i++) {
                var tuple = cookies[i].split('=')
                if (tuple[0] != 'userinfo') continue;
                if (tuple[1].length == 0) break; // no cookie value so don't bother
                var user = JSON.parse(atob(tuple[1]));
                //console.log("***** PRE:", tuple[1])
                //console.log("***** USER:", user)
                this.$dispatch('user-info', user)
                break
            }
        },
    }
})

var ipload = {
    methods: {
        loadData: function() {
            console.log("let all us load our data!")
            var self = this,
                 url = networkURL;
            if (self.STI > 0) {
                url +=  "?STI=" + self.STI;
                url += "&debug=true"
            }

            fetchData(url, function(data) {
                if (data) {
                    self.rows = data
                    console.log("loaded", data.length, "ip records")
                }
            })
        },
    }
}



var ipList = Vue.component('ip-list', {
    template: '#tmpl-ip-list',
    mixins: [ipload, pagedCommon, commonListMIX],
    created: function(ev) {
        console.log('ip list created!');
        this.loadData()
        this.title = "IP Addresses in Use"
    },
    data: function() {
        return {
            title: 'default title',
            rows: [],
            columns: [
                "Site",
                "Type",
                "Host",
                "Hostname",
                "IP",
                "Note"
            ],
            sortKey: '',
            sortOrders: [],
            Host: '',
            site: 'blah',
            STI: 1,
            IPT: 0,
            searchQuery: '',
            sites: [],
            typelist: [], 

            // TODO: kindlist should be populated from device_types
            hostlist: [
               '',
                'VM',
                'Server',
                'Switch',
            ],
        }
    },
    route: { 
          data: function (transition) {
            var self = this;
            console.log('server list promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(), 
                getIPTypes(),
           ]).then(function (data) {
              console.log('server list promises returning. site label:', self.site, 'STI:', self.STI)
              var types = data[1];
              types.unshift({IPT:0, Name: 'All'})
              return {
                sites: data[0],
                typelist: types,
              }
            })
          }
    },
    methods: {
        linkable: function(key) {
            return (key == 'Hostname')
        },
        linkpath: function(entry, key) {
            if (entry.Host != 'VM') {
                return '/device/edit/' + entry['ID']
            }
            return '/vm/edit/' + entry['ID']
        },
    },
    filters: {
        ipFilter: function(data) {
            if (! this.IPT && ! this.Host) {
                return data
            }

            var self = this;
            return data.map(function(obj) {
                if (self.IPT == obj.IPT && ! self.Host) {
                    return obj
                }
                if (self.Host == obj.Host && ! self.IPT) {
                    return obj
                }
                if (self.Host == obj.Host && self.IPT == obj.IPT) {
                    return obj
                }
            });
        },
    },
    events: {
        'ip-reload': function(msg) {
            console.log("reload those IPs!!!!!: ", msg)
        }
    },
    watch: {
        'STI': function(x) {
            this.loadData()
        }
    },
})


//
// IP TYPES
//
var ipTypes = Vue.component('ip-types', {
    template: '#tmpl-ip-types',
    mixins: [pagedCommon],
    data: function() {
        return {
            data: [],
            columns: [
                "Name",
                "Multi"
            ],
        }
    },
    route: { 
        data: function (transition) {
            var self = this;
            return Promise.all([
                getIPTypes(), 
            ]).then(function (data) {
                return {
                    data: data[0],
                }
            })
        }
    },
    methods: {
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/ip/type/edit/' + entry['IPT']
        }
    },
    watch: {
        'STI': function(x) {
            this.loadData()
        }
    },
})


var iptypeEdit = Vue.component('iptype-edit', {
    template: '#tmpl-iptype-edit',
    data: function() {
        return {
            IPType: new(IPType)
        }
    },
    route: {
        data: function (transition) {
            if (transition.to.params.IPT > 0) {
                return {
                    IPType: getIPTypes(transition.to.params.IPT)
                }
            }
            var ipType = new(IPType);
            ipType.IPT = 0
            return {
                IPType: ipType
            }
        }
    },
    methods: {
        newname: function() {
            console.log('my name is:', this.IPType.Name)
        },
        saveSelf: function() {
            var data = this.IPType;
            var id = data.IPT;
            var url = iptypesURL;
            if (id > 0) {
                postIt(url + id + "?debug=true", data, this.showList, 'PATCH')
            } else {
                postIt(url + id + "?debug=true", data, this.showList)
            }
        },
        deleteSelf: function() {
        },
        showList: function() {
            router.go('/ip/types')
        },
    },
})


//
// User List
//

var UserList = Vue.component('user-list', {
    template: '#tmpl-user-list',
    mixins: [pagedCommon],
    data: function() {
        return {
            columns: ['Login', 'First', 'Last', 'Access'],
            rows: [],
            url: userURL,
        }
    },
    created: function() {
        this.loadData()
    },
    methods: {
        loadData: function() {
            var self = this;

            fetchData(this.url, function(data) {
                if (data) {
                    self.rows = data
                    console.log("loaded", data.length, "ip records")
                }
            })
        },
        linkable: function(key) {
            return (key == 'Login')
        },
        linkpath: function(entry, key) {
            return '/user/edit/' + entry['USR']
        }
    }
})

//
// USER EDIT
//

var userEdit = Vue.component('user-edit', {
    template: '#tmpl-user-edit',
    mixins: [editVue],
    data: function() {
        return {
            User: new(User),
            dataURL: userURL,
            listURL: '/user/list',
            levels: [
                {Level:0, Label: 'User'},
                {Level:1, Label: 'Editor'},
                {Level:2, Label: 'Admin'},
            ],
        }
    },
    methods: {
        myID: function() {
            return this.User.USR
        },
        myself: function() {
            return this.User
        },
        loadSelf: function () {
            var self = this;
            var id = this.$route.params.USR;
            if (id > 0) {
                var url = this.dataURL + id;

                fetchData(url, function(data) {
                    self.User.Load(data);
                })
            }
        },
    },
})



var ipMIX = {
    props: ['iptypes', 'ports'],
    data: function() {
        return {
            newIP: '',
            newIPT: 0,
            newIFD: 0,
        }
    },
    computed: {
        ipAddDisabled: function() {
            return this.newIP.length == 0 || this.newIPT==0 || this.newIFD==0
        }
    },
    methods: {
        updateIP(i) {
            var row = this.rows[i]
            var iid = row.IID
            var ip = row.IP
            var ipt = row.IPT
            var ifd = row.IFD
            var data = {IFD:ifd, IID: iid, IPT: ipt, IPv4: ip}
            console.log('update IP:', ip, ' IID:', iid)
            postIt(ipURL + iid, data, null, 'PATCH')
            return false
        },
        deleteIP(i) {
            var self = this;
            var iid = this.rows[i].IID
            console.log("IP id:", iid)
            deleteIt(ipURL + iid, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    self.rows.splice(i, 1)
                }
            })
            return false
        },
        addIP: function() {
            var self = this;
            var data = {IFD: this.newIFD, IPT: this.newIPT, IPv4: this.newIP}
            console.log("we will add IP info:", data)
            postIt(ipURL + '?debug=true', data, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    var ip = JSON.parse(xhr.responseText);
                    ip.IP = ip.IPv4 // naming is inconsistent
                    self.rows.push(ip)

                    self.newIP = ''
                    self.newIPT = 0
                    self.newIFD = 0
                }
            })
            return false
        }
    }
}

var netgrid = childTable("network-grid", "#tmpl-network-grid", [ipMIX])

var interfaceMIX = {
    props: ['DID'],
    data: function() {
        return {
            newPort: '',
            newMgmt: false,
            newMAC: '',
            newSwitchPort: '',
            newCableTag: '',
        }
    },
    computed: {
        interfaceAddDisabled: function() {
            return this.newPort.length == 0 || this.newMAC.length == 0
        }
    },
    methods: {
        updateInterface(i) {
            var row = this.rows[i]
            var ifd = row.IFD
            postIt(ifaceURL + ifd, row, null, 'PATCH')
            return false
        },
        deleteInterface(i) {
            var self = this;
            var ifd = this.rows[i].IFD
            console.log("Iface id:", ifd)
            deleteIt(ifaceURL + ifd, function(xhr) {
                //console.log('del state:', xhr.readyState, 'status:', xhr.status)                
                if (xhr.readyState == 4) {
                    if (xhr.status == 200 || xhr.status == 201) {
                        self.rows.splice(i, 1)
                    } else {
                        var err = JSON.parse(xhr.responseText)
                        console.log('=============> ERROR:', err)
                    }
                }
            })
            return false
        },
        addInterface: function(ev) {
            var self = this
            var data = {
                DID: this.DID,
                Port: this.newPort,
                Mgmt: this.newMgmt,
                MAC: this.newMAC,
                SwitchPort: this.newSwitchPort,
                CableTag: this.newCableTag,
            }
            console.log("we will add interface info:", data)
            postIt(ifaceURL + '?debug=true', data, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    var iface = JSON.parse(xhr.responseText)
                    self.rows.push(iface)

                    self.newPort = ''
                    self.newMgmt = ''
                    self.newMAC = ''
                    self.newSwitchPort = ''
                    self.newCableTag = ''
                }
            })
            return false
        }
    }
}

var ifacegrid = childTable("interface-grid", "#tmpl-interface-grid", [interfaceMIX])

var uniqueInterfaces = function(data) {
    var ports = {}
    for (var i=0; i<data.length; i++) {
        var port = data[i]
        if (! (port.IFD in ports)) {
            console.log('IFD:', port.IFD)
            ports[port.IFD] = port

        }
    }
    var rep = [];
    for (var ifd in ports) {
        rep.push(ports[ifd])
    }
    return rep
}

//
// Device Edit
//
var deviceEditVue = {
  data: function() {
      return {
            sites: [],
            device_types: [],
            tags: [],
            ipTypes: [],
            newIP: '',
            newIPD: 0,
            newIFD: 0,
            netColumns: ['IP', 'Type', 'Port'],
            ifaceColumns: ['Port', 'Mgmt', 'MAC', 'CableTag', 'SwitchPort'],
            Description: '',
            Device: new(Device),
        }
    },
    route: { 
          data: function (transition) {
            console.log('DEVICE ROUTE TRANS:',transition)
            console.log('route did:',this.$route.params.DID)
            return Promise.all([
                getSiteLIST(false), 
                getDeviceTypes(), 
                getTagList(),
                getIPTypes(),
                completeDevice(this.$route.params.DID), 
           ]).then(function (data) {
              return {
                sites: data[0],
                device_types: data[1],
                tags: data[2],
                ipTypes:  data[3],
                Device: data[4],
               }
            })
          }
    },
    methods: {
        saveSelf: function(event) {
            var device = this.Device;
            delete device['racks'];
            delete device['interfaces'];
            delete device['ips'];

            if (this.Device.DID == 0) {
                console.log('save new device');
                postIt(deviceURL + "?debug=true", device, this.showList)
                return
            }
            console.log('update device id: ' + this.Device.DID);
            postIt(deviceURL + this.Device.DID + "?debug=true", device, this.showList, 'PATCH')
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            deleteIt(deviceURL + this.Device.DID, this.showList)
        },
        showList: function(ev) {
            //this.$route.router.go(window.history.back())
            router.go('/device/list')
        },
        loadRacks: function () {
             var self = this;
            console.log("RACK URL:", rackURL + "?STI=" + self.Device.STI)
             fetchData(rackURL + "?STI=" + self.Device.STI, function(data) {
                 self.racks = data
             })
        },
        loadTags: function () {
             var self = this;
             fetchData(tagURL, function(data) {
                 self.tags = data
             })
        },
        portLabel: function(ipinfo) {
            if (this.Device.DevType === 'server') {
                return ipinfo.Mgmt ? 'IPMI' : 'Eth' + ipinfo.Port
            } 
            return (ipinfo.Mgmt ? 'Mgmt' : 'Port') + ipinfo.Port
        },
        getMacAddr: function(ev) {
            //var url = '/dcman/data/server/discover/' + this.Server.IPIpmi;
            var url = 'http://10.100.182.16:8080/dcman/data/server/discover/' + this.Server.IPIpmi;
            var self = this
            fetchData(url, function(data) {
                self.Device.MacPort0 = data.MacEth0
                console.log("MAC DATA:", data)
             })
             ev.preventDefault();
            return false;
        },

    },
}

var deviceEdit = Vue.component('device-edit', {
    template: '#tmpl-device-edit',
    mixins: [deviceEditVue],
})


var deviceAddMIX = {
    route: { 
          data: function (transition) {
            var device = new(Device);
            device.DID    = 0;
            device.Height = 1;
            device.TID    = 1;
            device.Rack   = 0;
            device.STI    = parseInt(this.$route.params.STI);
            device.RID    = parseInt(this.$route.params.RID);
            device.RU     = parseInt(this.$route.params.RU);
            return {
                Device: device
            }
          },
    },
}

var deviceAdd = Vue.component('device-add', {
    template: '#tmpl-device-edit',
    mixins: [deviceEditVue, deviceAddMIX],
})

//
// VM IPs
//
var vmIpMIX = {
    props: ['VMI'],
    data: function() {
        return {
            newIP: '',
            newIPT: 0,
            types: [],
            rows: [],
        }
    },
    created: function () {
        console.log('CREATED VMI:', this.VMI)
        this.loadSelf()
    },
    computed: {
        ipAddDisabled: function() {
            return this.newIP.length == 0 || this.newIPT==0 || this.newIFD==0
        }
    },
    methods: {
        loadSelf: function() {
            var self = this;
            console.log('MY VMI:', this.VMI)
            var url = ipURL + '?VMI=' + this.VMI;
            fetchData(url, function(data) {
                 self.rows = data
            })
            fetchData(iptypesURL, function(data) {
                 self.types = data
                 console.log("IP TYPES:", data)
            })
        },
        updateIP(i) {
            var row = this.rows[i]
            var iid = row.IID
            var ip = row.IPv4
            var ipt = row.IPT
            var data = {VMI:this.VMI, IID: iid, IPT: ipt, IPv4: ip}
            console.log('update IP:', ip, ' IID:', iid)
            postIt(ipURL + iid, data, null, 'PATCH')
            return false
        },
        deleteIP(i) {
            var self = this;
            var iid = this.rows[i].IID
            console.log("IP id:", iid)
            deleteIt(ipURL + iid, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    self.rows.splice(i, 1)
                }
            })
            return false
        },
        addIP: function() {
            var self = this;
            var data = {VMI: this.VMI, IPT: this.newIPT, IPv4: this.newIP}
            postIt(ipURL + '?debug=true', data, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    self.rows.push(data)
                    self.newIP = ''
                    self.newIPT = 0
                }
            })
            return false
        }
    },
    watch: {
        'VMI': function() {
            this.loadSelf()
        }
    }
}

var vmips = Vue.component('vm-ips', {
    template: '#tmpl-vm-ips',
    mixins: [vmIpMIX],
})


//
// VM Edit
//
var vmEditVue = {
    data: function() {
        return {
            url: vmViewURL,
            STI: 0,
            sites: [],
            racks: [],
            tags: [],
            ipTypes: [],
            ipRows: [],
            Description: '',
            VMI: parseInt(this.$route.params.VMI),
            VM: new(VM),
        }
    },
    route: { 
          data: function (transition) {
            return {
                VM: getVM(this.$route.params.VMI)
            }
          },
    },
    methods: {
        saveSelf: function(event) {
            console.log('send update event: ' + event);
            postIt(this.url + this.VM.VMI + "?debug=true", this.VM, this.showList, 'PATCH')
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            postIt(this.url + this.VM.VMI, null, this.showList, 'DELETE')
        },
        showList: function(ev) {
            router.go('/vm/list')
        },
    },
}

var vmEdit = Vue.component('vm-edit', {
    template: '#tmpl-vm-edit',
    mixins: [vmEditVue, siteMIX],
})

// Base APP component, this is the root of the app
var App = Vue.extend({
    data: function(){
        return {
            myapp: {
                auth: {
                    loggedIn: false,
                    user: {
                        name: null, 
                        admin: 0,
                    }
                },
            },
        }
    },
    methods: {
        fresher: function(ev) {
            console.log("the fresh maker!")
            this.$broadcast('ip-reload', 'please')
        },
    },
    events: {
        'server-found': function (ev) {
            console.log('app reload event:', ev)
            this.$broadcast('server-reload', 'gotcha!')
        },
        'user-info': function (user) {
            //console.log('*** user-info event:', user)
            this.myapp.auth.user.name = user.username;
            this.myapp.auth.loggedIn = true;
        },
        'user-auth': function (user) {
            console.log('*** user auth event:', user)
            this.myapp.auth.user.name = user.Login;
            this.myapp.auth.user.admin = user.Level;
            window.user_apikey = user.APIKey
            userInfo = user;
            this.myapp.auth.loggedIn = true;
            //user_apikey = user.apikey;
        },
        'logged-out': function () {
            console.log('*** logged out event')
            this.myapp.auth.user.name = null
            this.myapp.auth.user.admin = 0
            this.myapp.auth.loggedIn = false
            window.user_apikey = ''
            fetchData('/dcman/api/logout')
        },
        'search-again': function(text) {
            // relay event from navbar search
            this.$broadcast('search-again', text)
        },
    },
})


//
// DEVICE LIST
//

var deviceList = Vue.component('device-list', {
    template: '#tmpl-device-list',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
      console.log('device data returning')
      return {
          STI: 1,
          RID: 0,
          sites: [],
          racks: [],
          site: 'blah',
          searchQuery: '',
          pagerows: 10,
          rows: [],
          columns: [
               "Site",
               "Rack",
               "RU",
               "Hostname",
               "IPs",
               "Mgmt",
               "Tag",
               "Profile",
               "SerialNo",
               "AssetTag",
               "Assigned",
               "Note",
            ],
        }
    },
    filters: {
        rackFilter: function(data) {
            if (! this.RID) return data;

            var self = this;
            return data.map(function(obj) {
                 if (obj.RID == self.RID) return obj
            });
        }
    },
    route: { 
          data: function (transition) {
            var self = this;
            console.log('device list promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(), 
                getDeviceLIST(self.STI, '?sti='), 
                getRack(self.STI, '?sti='), 
           ]).then(function (data) {
                console.log('device list promises returning')
                var racks =  data[2]
                racks.unshift({RID:0, Label:'All'})
             return {
                sites: data[0],
                rows: data[1],
                racks: data[2],
              }
            })
          }
    },
    methods: {
        reload: function() {
            var self = this;
            getDeviceLIST(self.STI, '?sti=').then(function(devices) {
                self.rows = devices
            })
            getRack(self.STI, '?sti=').then(function(racks) {
                self.racks = racks
            })
        },
        canLink: function(column) {
            return column === 'Hostname'
        },
        linkFN: function(entry, key) {
            if (key == 'Hostname') return '/device/edit/' + entry['DID']
        }
    },
    watch: {
    'STI': function(val, oldVal){
            mySTI = val
            this.reload()
        },
    },
})


//
// VLAN List
//

var vlanList = Vue.component('vlan-list', {
    template: '#tmpl-vlan-list',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            dataURL: '/dcman/api/vlan/view/',
            listURL: '/vlan/list',
            STI: 0,
            sites: [],
            searchQuery: '',
            data: [],
            columns: [
                "Site",
                "Name",
                "Profile",
                "Gateway",
                "Route",
                "Netmask",
            ]
         }
    },
    created: function () {
        this.loadSelf()
        var self = this;
        getSiteLIST(true).then(function(data) {
            self.sites = data
        })
    },
    methods: {
        loadSelf: function () {
            var self = this;
            var url = this.dataURL;
            fetchData(url, function(data) {
                self.data = data
            })
        },
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            console.log('vlan entry:', entry)
            return '/vlan/edit/' + entry['VLI']
        },
    },
    filters: {
        siteFilter: function(data) {
            console.log('filter for:',this.STI)
            if (! this.STI) return data;

            var self = this;
            return data.map(function(obj) {
                 if (obj.STI == self.STI) return obj
            });
        }
    }
})


//
// VLAN Edit
//

var vlanEdit = Vue.component('vlan-edit', {
    template: '#tmpl-vlan-edit',
    mixins: [editVue, commonListMIX, siteMIX],
    data: function() {
        return {
            sites: [],
            VLAN: new(VLAN),
            dataURL: '/dcman/api/vlan/view/'
        }
    },
    route: { 
          data: function (transition) {
            return {
                VLAN: getVLAN(this.$route.params.VLI)
            }
          },
    },
    methods: {
        myID: function() {
              return this.VLAN.VLI
        },
        myself: function() {
              return this.VLAN
        },
        showList: function(ev) {
              router.go('/vlan/list')
        },
    },
})



//
// IP Reserve
//
var ipReserveVue = {
    data: function() {
        return {
            sites: [],
            vlans: [],
            From: '',
            To: '',
            Note: '',
            Network: '',
            Netmask: '',
            Max: '',
            STI: 0,
            VLI: 0,
            minIP32: 0,
            maxIP32: 0,
            VLAN: new(VLAN),
            dataURL: '/dcman/api/vlan/view/'
        }
    },
    computed: {
        disableReserve: function() {
            if (this.STI == 0) return true
            if (this.VLI == 0) return true
            if (this.From.length == 0 || ! validIP(this.From)) return true 
            if (this.To.length == 0 || ! validIP(this.To)) return true 
            var from = ip32(this.From)
            var to = ip32(this.To)
            if (from < this.minIP32 || from > this.maxIP32) return true
            if (to < this.minIP32 || to > this.maxIP32) return true
            if (to < from) return true
            return false
        }
    },
    methods: {
        showList: function() {
            router.go('/ip/list')
        },
        reserveIPs: function() {
            var url = '/dcman/api/network/ip/range';
            var data = {
                From: this.From,
                To: this.To,
            }
            posty(url, data).then(function(ips) {
                console.log("IPS IN RANGE:", ips)
            })
        },
        checkFrom: function() {
            if (! validIP(this.From)) {
                alert('Invalid IP:', this.From)
            }
        },
        checkTo: function() {
            if (! validIP(this.To)) {
                alert('Invalid IP:', this.To)
            }
        },
    },
    watch: {
        STI: function() {
            var self = this;
            var url = vlanURL + "?STI=" + this.STI
            fetchData(url, function(data) {
                console.log('loaded vlan cnt:', data.length)
                self.vlans = data
            })
        },
        VLI: function() {
            console.log('VLI:', this.VLI, 'cnt:', this.vlans.length)
            for (var i=0; i < this.vlans.length; i++) {
                var vlan = this.vlans[i]
                console.log('i:', i, 'vlan:', vlan.VLI)
                if (vlan.VLI == this.VLI) {
                    var full = (1 << 32) - 1
                    var mask = ip32(vlan.Netmask)
                    var net = ip32(vlan.Gateway)
                    var min_ip = (net & mask) + 1
                    var max_ip = (net | ~mask) - 1
                    /*
                    this.Netmask = vlan.Netmask
                    this.Network = ipv4(min_ip)
                    this.Max = ipv4(max_ip)
                    */
                    this.minIP32 = min_ip
                    this.maxIP32 = max_ip
                    this.Network = ipv4(min_ip) + ' - ' + ipv4(max_ip)
                    break
                }
            }
        }
    }
}

var ipReserved = Vue.component('ip-reserve', {
    template: '#tmpl-ip-reserve',
    mixins: [ ipReserveVue, siteMIX],
})


//
// VM LIST
//

var vmList = Vue.component('vm-list', {
    template: '#tmpl-vm-list',
    mixins: [pagedCommon, siteMIX, commonListMIX],
    data: function() {
        return {
            STI: 1,
            sites: [],
            site: 'blah',
            searchQuery: '',
            data: [],
            columns: [
                 "Site",
                 "Server",
                 "Hostname",
                 "Profile",
                 "Note",
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
    loadRacks: function () {
         var self = this,
              url = rackURL + "?STI=" + self.STI;

         fetchData(url, function(data) {
             if (data) {
                 data.unshift({RID:0, Label:''})
                 self.racks = data
             }
         })
    },
    loadSelf: function () {
         var self = this

         var url = vmViewURL;
         if (self.STI > 0) {
             url += "?sti=" + self.STI
         } 
         fetchData(url, function(data) {
             self.data = data
         })
         self.loadRacks()
        
        },
        linkable: function(key) {
            return (key == 'Hostname' || key == 'Server')
        },
        linkpath: function(entry, key) {
            if (key == 'Server') return '/device/edit/' + entry['DID']
            if (key == 'Hostname') return '/vm/edit/' + entry['VMI']
        }
  },
  watch: {
    'STI': function(val, oldVal){
            this.loadSelf()
        },
    },
})



//
// Inventory
//

var updateQty = function(id, qty) {
    data = {ID:parseInt(id), Qty:parseInt(qty)}
    if (data.Qty < 0) {
        alert("quantity cannot be negative");
        return;
    }
    console.log("Update Data:",data)
    postIt(partURL + id, data, null, 'PATCH')
}

var addedItem = function(xref) {
    console.log("results:" + xref)
}



var partInventory = Vue.component('part-inventory', {
    template: '#tmpl-part-inventory',
    mixins: [pagedCommon, commonListMIX, siteMIX],
    data: function() {
        return {
            showgrid: true,
            DID: 0,
            STI: 4,
            PID: 0,
            available: [],
            sites: [],
            hostname: '',
            other: '',
            searchQuery: '',
            partData: [],
            columns: ['Site', 'Description', 'PartNumber', 'Mfgr', 'Qty', 'Price'],
        }
    },
    route: {
        data: function (transition) {
            return {
                partData: getInventory(this.STI, '?sti='),
                sites: getSiteLIST(true), 
            }
        }
    },
    methods: {
        updated: function(event) {
            console.log('the event: ' + event)
        },
        linkable: function(key) {
            return (key == 'Description')
        },
        linkpath: function(entry, key) {
            return '/part/use/' + entry['STI'] + '/' + entry['KID']
        },
    },
    watch: {
        'STI': function(val, oldVal){
             var self = this;
             getInventory(this.STI, '?sti=').then(function(data) {
                 self.partData = data
             })
        },
    },
})



var partUse = Vue.component('part-use', {
    template: '#tmpl-part-use',
    data: function() {
        return {
            badHost: false,
            DID: 0,
            STI: 4,
            PID: 0,
            available: [],
            sites: [],
            hostname: '',
            other: '',
            searchQuery: '',
            partData: [],
        }
    },
    created: function () {
        this.loadData()
    },
    methods: {
        showList: function() {
            router.go('/part/inventory')
        },
        loadData: function () {
            var self = this
            var kid = this.$route.params.KID;
            var sti = this.$route.params.STI;
            var url = partURL + "?unused=1&bad=0&kid=" + kid + "&sti=" + sti
            fetchData(url, function(data) {
                self.available = data
            })
        },
        thisPart: function(ev) {
          //alert("ok!")
          var pid = document.getElementById("PID").value
          var part = {
            PID: parseInt(pid),
            STI: this.STI,
            DID: this.DID,
            Unused: false,
          }
          postIt(partURL + pid, part, null, "PATCH")
          this.showList()
      },
        findhost: function() {
            if (this.hostname.length === 0) {
                this.badHost = false
                return
            }
            var self = this;
            var url = deviceURL + "?hostname=" + this.hostname;
            getJSON(url).then(function(hosts) {
                if (hosts && hosts.length == 1) {
                    self.RMA.DID = hosts[0].DID
                    self.badHost = false;
                } else {
                    self.badHost = true;
                }
            })
        },
    },
})


//
// Parts
//

var partList = Vue.component('part-list', {
    template: '#tmpl-part-list',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            showgrid: true,
            isgood: true,
            isbad: false,
            DID: 0,
            STI: 0,
            PID: 0,
            available: [],
            sites: [],
            hostname: '',
            other: '',
            searchQuery: '',
            ktype: 1,
            kinds: [
                {id: 1, name: "All Parts"},
                {id: 2, name: "Good Parts"},
                {id: 3, name: "Bad Parts"},
            ],
            rows: [],
            columns: [
               "Site",
               "Hostname",
               "Serial",
               "PartType",
               "PartNumber",
               "Description",
               "Vendor",
               "Mfgr",
               "Price",
               "Bad",
            ]
        }
    },
    route: { 
          data: function (transition) {
            var self = this;
            var STI = parseInt(transition.to.params.STI);
            console.log('part list promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(true), 
                (STI > 0 ? getPart(STI, '?sti=') : getPart()), 
           ]).then(function (data) {
              console.log('part list promises returning')
             return {
                sites: data[0],
                rows: data[1],
                STI: STI,
              }
            })
        }
    },
    methods: {
      findhost: function(ev) {
          var self = this;
          console.log("find hostname:",this.hostname);
          fetchData("api/server/hostname/" + this.hostname, function(data, status) {
               var enable = (status == 200);
               buttonEnable(document.getElementById('use-btn'), enable)
               self.DID = enable ? data.ID : 0;
            })
        },
        newPart: function(ev) {
            var id = parseInt(ev.target.id.split('-')[1]);
        },
        linkable: function(key) {
            return (key == 'Description')
        },
        linkpath: function(entry, key) {
            return '/part/edit/' + entry['PID']
        }
    },
    watch: {
        'STI': function(newVal,oldVal) {
            router.go('/part/list/' + newVal)
        }
    },
    filters: {
        partFilter: function(data) {
            if (this.ktype == 1) {
                return data
            }
            var self = this;
            return data.map(function(obj) {
                if (self.ktype == 2 && ! obj.Bad) {
                    return obj 
                }
                if (self.ktype == 3 && obj.Bad) {
                    return obj
                }
            });

        },
    }
})


//
// PART TYPES
//

var partTypes = Vue.component('part-types', {
    template: '#tmpl-part-types',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            searchQuery: '',
            rows: [],
            columns: [
               "Name",
            ]
        }
    },
    route: { 
          data: function (transition) {
            return {
              rows: getPartTypes(),
            }
          }
    },
    methods: {
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/part/type/edit/' + entry['PTI']
        }
    }
})

//
// RMAs
//

// register the grid component

var rmaListMIX = {
    props: ['kfilter'],
    data: function () {
        var sortOrders = {}
        this.columns.forEach(function (key) {
            sortOrders[key] = 1
        })
        return {
            sortKey: '',
            sortOrders: sortOrders
        }
    },
    methods: {
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
        linkable: function(key) {
            switch(key) {
                case 'Description': return true;
                case 'Hostname': return true;
            }
            return false;
            //return (key == 'Description')
        },
        linkpath: function(entry, key) {
            //return '/rma/edit/' + entry['RMD']
            switch(key) {
                case 'Description': return '/rma/edit/' + entry['RMD']
                case 'Hostname': return '/device/edit/' + entry['DID']
            }
        },
        subFilter: function(a, b, c) {
            if (this.kfilter == 1) {
                return a
            }
            if (this.kfilter == 2 && ! a.Closed) {
                return a
            }
            if (this.kfilter == 3 && a.Closed) {
                return a
            }
        },
    },
}


var rmaListVue = {
    data: function() {
        return {
            STI: 4,
            sites: [],
            rmas: [],
            searchQuery: '',
            ktype: 1,
            gridColumns: [
                "RMD",
                "Description",
                "Hostname",
                //"ServerSN",
                "PartSN",
                "VendorRMA",
                "Jira",
                "Created",
                "Shipped",
                "Received",
                "Closed",
            ],
            kinds: [
                {id: 1, name: "All RMAs"},
                {id: 2, name: "Open RMAs"},
                {id: 3, name: "Closed RMAs"},
            ],
    }
  },
  created: function () {
      this.loadSelf()
  },
  methods: {
    loadSelf: function () {
         var self = this;
         fetchData(rmaviewURL + "?STI=" + self.STI, function(data) {
             self.rmas = data
         })
    },
  },
  watch: {
    'STI': function(val, oldVal){
            this.loadSelf()
        },
    },
}

var rg = childTable("rma-grid", "#tmpl-base-table", [rmaListMIX])

var rList = Vue.component('rma-list', {
    template: '#tmpl-rma-list',
    mixins: [rmaListVue, siteMIX],
})

var foundListMIX = {
  data: function () {
      var sortOrders = {}
      this.columns.forEach(function (key) {
          sortOrders[key] = 1
      })
      return {
          sortKey: '',
          sortOrders: sortOrders
      }
    },
    methods: {
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
  },
}

//
// VENDOR LIST
//

var vendorListVue = {
    data: function() {
        return {
            sites: [],
            searchQuery: '',
            rows: [],
            columns: [
                'Name',
                //'WWW',
                'Phone',
/*
                'Address',
                'City',
                'State',
                'Country',
                'Postal',
*/
                'Note',
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            var self = this;
            fetchData(vendorURL, function(data) {
                self.rows = data
            })
        },
    },
}

var vendorListMIX = {
    methods: {
/*
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
*/
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/vendor/edit/' + entry['VID']
        },
    }
}
var vendg = childTable("vendor-grid", "#tmpl-base-table", [vendorListMIX])

var vendorList = Vue.component('vendor-list', {
    template: '#tmpl-vendor-list',
    mixins: [vendorListVue],
})

var vendorEditVue = {
    data: function() {
        var vendor = new(Vendor);
        vendor.VID = 0
        return {
            Vendor: vendor,
            dataURL: vendorURL,
            listURL: '/vendor/list',
        }
    },
    methods: {
        myID: function() {
            return this.Vendor.VID
        },
        myself: function() {
            return this.Vendor
        },
        loadSelf: function () {
            var self = this;
            var id = this.$route.params.VID;
            if (id > 0) {
                var url = this.dataURL + id;

                fetchData(url, function(data) {
                    self.Vendor.Load(data);
                })
            }
        },
    },
}

var vendorEdit = Vue.component('vendor-edit', {
    template: '#tmpl-vendor-edit',
    mixins: [vendorEditVue, editVue],
})

//
// MFGR LIST
//

var mfgrListVue = {
    data: function() {
        return {
            sites: [],
            searchQuery: '',
            rows: [],
            columns: [
                'Name',
                'Note',
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            var self = this;
            fetchData(mfgrURL, function(data) {
                self.rows = data
            })
        },
    },
}

var mfgrListMIX = {
    methods: {
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/mfgr/edit/' + entry['MID']
        },
    }
}
var mfgrg = childTable("mfgr-grid", "#tmpl-base-table", [mfgrListMIX])

var mfgrList = Vue.component('mfgr-list', {
    template: '#tmpl-mfgr-list',
    mixins: [mfgrListVue],
})

var mfgrEditVue = {
    data: function() {
        var mfgr = new(Mfgr);
        mfgr.MID = 0
        return {
            Mfgr: mfgr,
        }
    },
    methods: {
        showList: function(xhr) {
            router.go('/mfgr/list')
        },
        deleteSelf: function() {
            deleteIt(mfgrURL + this.Mfgr.MID, this.showList)
        },
        saveSelf: function()  {
            saveMe(mfgrURL, this.Mfgr, this.Mfgr.MID, this.showList)
        } 
    },
    route: { 
          data: function (transition) {
            //console.log('part list promises starting for STI:', self.STI) 
            return Promise.all([
                getMfgr(transition.to.params.MID)
           ]).then(function (data) {
             return {
                Mfgr: data[0],
              }
            })
        }
    },
}

var mfgrEdit = Vue.component('mfgr-edit', {
    template: '#tmpl-mfgr-edit',
    mixins: [mfgrEditVue],
})

//
// PART EDIT
//
var partEditVue = {
    data: function() {
        var part = new(Part);
        part.PID = 0;
        part.DID = 0;
        part.STI = 0;
        part.PTI = 0;
        part.VID = 0;
        part.Bad = false;
        part.Used = false;
       return {
            badHost: false,
            sites: [],
            types: [],
            vendors: [],
            Part: part,
       }
    },
    computed: {
        disableSave: function() {
            console.log("part PTI:", this.Part.PTI)
            if (this.Part.PTI == 0) {
                return (this.Part.STI == 0 || this.Part.PTI == 0 || this.Part.Description.length == 0)
            }
        },
        badPrice: function() {
            return false
        }
    },
    created: function () {
        this.loadPart()
    },
    route: { 
          data: function (transition) {
            //var userId = transition.to.params.userId
            return {
              sites: getSiteLIST(), //sitePromise,
              types: getPartTypes(),
              vendors: getVendorList().then(function(list) {
                  list.unshift({VID:0, Name:''})
                    return list
              }),
            }
          }
    },
    methods: {
        showList: function(ev) {
            router.go('/part/list/' + this.Part.STI)
        },
        validprice: function() {
        },
        saveSelf: function(event) {
            this.Part.Price = parseFloat(this.Part.Price)
            this.Part.Cents = Math.round(this.Part.Price * 100)
            console.log('save part price: ' + this.Part.Price, 'cents:', this.Part.Cents)
            var url = partURL;
            if (this.Part.PID > 0) {
                url += this.Part.PID
                url += "?debug=true"
                postIt(url, this.Part, this.showList, 'PATCH')
            } else {
                url += "?debug=true"
                postIt(url, this.Part, this.showList)
            }
        },
        doRMA: function(ev) {
            router.go('/rma/create/' + this.Part.PID)
        },
        loadPart: function () {
             var id = this.$route.params.PID;
             if (id > 0) {
                 var url = partURL + id;

                 self = this
                 fetchData(url, function(data) {
                     self.Part.Load(data);
                 })
             }
        },
        findhost: function() {
            if (this.Part.Hostname.length === 0) {
                this.Part.DID = 0
                this.badHost = false
                return
            }
            var self = this;
            var url = deviceURL + "?hostname=" + this.Part.Hostname;
            getJSON(url).then(function(hosts) {
                if (hosts && hosts.length == 1) {
                    self.Part.DID = hosts[0].DID
                    self.badHost = false;
                } else {
                    self.badHost = true;
                }
            })
        },
    },
}


var pEdit = Vue.component('part-edit', {
    template: '#tmpl-part-edit',
    mixins: [partEditVue],
})


//
// RMA EDIT
//
var rmaCommon = {
    data: function() {
        var rma = new(RMA);
        rma.RMD = 0
        rma.NewPID = 0
        rma.OldPID = 0
        rma.UID = 0
        return {
            badHost: false,
            dataURL: rmaviewURL,
            RMA: rma
        }
    },
    methods: {
        saveSelf: function(event) {
            if (this.RMA.RMD > 0) {
                postIt(rmaviewURL + this.RMA.RMD + "?debug=true", this.RMA, this.showList, 'PATCH')
            } else {
                postIt(rmaviewURL+ "?debug=true", this.RMA, this.showList)
            }
        },
        showList: function() {
            router.go('/rma/list')
        },
        findhost: function() {
            if (this.RMA.Hostname.length === 0) {
                this.RMA.DID = 0
                this.badHost = false
                return
            }
            var self = this;
            var url = deviceURL + "?hostname=" + this.RMA.Hostname;
            getJSON(url).then(function(hosts) {
                if (hosts && hosts.length == 1) {
                    self.RMA.DID = hosts[0].DID
                    self.badHost = false;
                } else {
                    self.badHost = true;
                }
            })
        },
    },
}


var rmaEditVue = {
    route: { 
        data: function (transition) {
            return {
                RMA: getRMA(this.$route.params.RMD),
            }
        }
    },
    methods: {
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            postIt(this.dataURL + this.myID(), null, this.showList, 'DELETE')
        },
    },
}

var rEdit = Vue.component('rma-edit', {
    template: '#tmpl-rma-edit',
    mixins: [rmaCommon, rmaEditVue],
})

//
// RMA CREATE
//

var rmaCreateVue = {
    route: { 
        data: function (transition) {
            var part = getPart(this.$route.params.PID);
            var self = this;
            return {
                RMA: getPart(this.$route.params.PID).then(function(part) {
                        // TODO: this is goofy -- where to initialize RMA?
                        var rma = new(RMA);
                        rma.RMD = 0
                        rma.DID = part.DID
                        rma.VID = part.VID
                        rma.STI = part.STI
                        rma.OldPID = part.PID
                        rma.NewPID = 0
                        rma.Description = part.Description
                        rma.PartNumber = part.PartNumber
                        rma.Hostname = part.Hostname
                        rma.PartSN = part.Serial
                        return(rma)
                }),
            }
        },
    },
}

var rCreate = Vue.component('rma-create', {
    template: '#tmpl-rma-edit',
    mixins: [rmaCommon, rmaCreateVue],
})


//
// PART LOAD
//
var partLoadVue = {
    data: function() {
        return {
            Parts: '',
            sites: [],
            STI: 2,
        }
    },
    route: {
        data: function (transition) {
            return {
                sites: getSiteLIST()
            }
        }
    },
    methods: {
        showList: function(ev) {
            router.go('/part/list/' + this.STI)
        },
        saveParts: function() {
            var partCol = function(col) {
                switch (col) {
                    case "Item":            return "PartType";
                    case "Part Number":     return "PartNumber";
                    case "SKU":             return col;
                    case "Description":     return col;
                    case "Manufacturer":    return "Mfgr";
                    case "Qty":             return col;
                    case "Price":           return col;
                }
                return ""
            }
            var parts = this.Parts.split("\n");
            var cols = {};
            for (var i=0; i < parts.length; i++) {
                var line = parts[i].split("\t");
                if (i === 0) {
                    for (var k=0; k < line.length; k++) {
                        console.log("COL:", line[k])
                        var col = partCol(line[k])
                        if (col.length > 0) {
                            cols[k] = col
                        }
                    }
                    console.log("COLS:", cols)
                    continue
                }
                var part = new(Part);
                part.PID = 0;
                part.KID = null;
                part.DID = null;
                part.VID = 0;
                part.STI = this.STI;
                part.Bad = false;
                part.Unused = true;
                for (var j in cols) {
                    var col = cols[j];
                    part[col] = line[j];
                }
                //console.log("PART:",part)
                var qty = parseInt(part["Qty"]);
                if (qty === 0) qty = 1;
                //console.log("Price was:", part.Price)
                if (part.Price) {
                    part.Price = part.Price.replace(/[^0-9.]*/g,'')
                    //console.log("Price fix:", part.Price)
                    part.Price = parseFloat(part.Price)
                    //console.log("Price now:", part.Price)
                    part.Cents = Math.round(part.Price * 100)
                    //console.log("Cents now:", part.Cents)
                } else {
                    part.Price = 0.0
                    part.Cents = 0
                }
                var url = partURL;
                url += "?debug=true"
                for (var j=0; j < qty; j++) {
                    postIt(url, part, function(xhr) {});
                }
            }
        },
    },
}

var ploadVue = Vue.component('part-load', {
    template: '#tmpl-part-load',
    mixins: [partLoadVue],
})

//
// TAGS
//

var tagEditVue = {
    props: ['columns', 'rows'],
    data: function () {
        var tag = new(Tag);
        tag.TID = 0
        return {
            tags: [],
            url: tagURL,
            tag: tag,
            sites: [],
        }
    },
    route: { 
          data: function (transition) {
            //var userId = transition.to.params.userId
            return {
              sites: getSiteLIST(), //sitePromise,
            }
          }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        showList: function() {
            router.go('/')
        },
        loadSelf: function () {
             var self = this;
             fetchData(this.url, function(data) {
                 self.tags = data
             })
        },
        deleteSelf: function(ev) {
            console.log("delete self...")
            if (! this.tag) {
                return
            }
            console.log('delete tag url: ' + this.url + this.tag.TID)
            postIt(this.url + this.tag.TID, null, function(data) {}, 'DELETE')
            console.log("delete tid:",this.tag.TID)
            for (var i=0; i < this.tags.length; i++) {
                console.log("i:",i,"tid:",this.tags[i].TID)
                if (this.tags[i].TID == this.tag.TID) {
                    console.log("deleting tag:", i, "of", this.tags.length)
                    //delete(this.tags[i])
                    this.tags.splice(i, 1)
                    break
                }
            }
            this.tag = new(Tag)
            this.tag.TID = 0
            ev.preventDefault()
        },
        saveSelf: function(event) {
            var self = this;
            //this.tag.TID = parseInt(this.tag.TID)
            //console.log('update tag event: ' + event);
            var saved = function(xreq) {
                if (xreq.readyState == 4) {
                     if (xreq.status != 201) {
                        alert("Oops: ("+xreq.status+") " + xreq.responseText);
                        return
                     }
                     self.tag = JSON.parse(xreq.responseText)
                    self.loadSelf()
                }
            }
            if (this.tag.TID > 0) {
                // postIt = function(url, data, fn, method) {
                var self = this
                var refresh = function() {
                    for (var i=0; i < self.tags.length; i++) {
                        if (self.tags[i].TID == self.tag.TID) {
                            self.tags[i].Name = self.tag.Name
                            break
                        }
                    }
                }
                postIt(this.url + this.tag.TID + "?debug=true", this.tag, refresh, 'PATCH')
            } else {
                postIt(this.url + "?debug=true", this.tag, saved)
            }
            //this.loadSelf()
        },
    },
    watch: {
        'tag.TID': function() {
            console.log('this tag:', this.tag)
            for (var i=0; i < this.tags.length; i++) {
                console.log('i:',i, "tag:", this.tag[i])
                if (this.tags[i].TID == this.tag.TID) {
                    this.tag.Name = this.tags[i].Name
                    return
                }
            }
            this.tag.Name = ''
        }
    },
}

var tagEdit = Vue.component('tag-edit', {
    template: '#tmpl-tag-edit',
    mixins: [tagEditVue],
})

//
// RACK Edit
//
var rackEditVue = {
    data: function() {
        return {
            sites: [],
            id: 'RID',
            name: 'Rack',
            Rack: new(Rack),
            dataURL: '/dcman/api/rack/view/',
            listURL: '/rack/list',
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        myID: function() {
              return this.Rack.RID
          },
          myself: function() {
              return this.Rack
          },
          showList: function(ev) {
              router.go('/rack/list')
          },
          loadSelf: function () {
               var self = this;

               var id = this.$route.params[this.id];
               console.log('loading rack ID:', id)
               if (id > 0) {
                   var url = this.dataURL + id;

                   fetchData(url, function(data) {
                       self.Rack.Load(data);
                   })
                 }
          },
    },
}

var rackEdit = Vue.component('rack-edit', {
    template: '#tmpl-rack-edit',
    mixins: [editVue, rackEditVue, siteMIX],
})

//
// RACK LIST
//

var rackList = Vue.component('rack-list', {
    template: '#tmpl-rack-list',
    mixins: [pagedCommon, siteMIX],
    data: function() {
        return {
            dataURL: '/dcman/api/rack/view/',
            STI: 1,
            sites: [],
            searchQuery: '',
            rows: [],
            site: 'blah',
            columns: [
               "Site",
               "Label",
               "VendorID",
               "RUs",
               "Note",
            ]
        }
    },

    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
             var self = this

             var url = this.dataURL;
             if (self.STI > 0) {
                 url += "?sti=" + self.STI
             }
             fetchData(url, function(data) {
                 self.rows = data
             })
        },
        linkable: function(key) {
            return (key == 'Label')
        },
        linkpath: function(entry, key) {
            if (key == 'Label') return '/rack/edit/' + entry['RID']
        },
  },

  watch: {
    'STI': function(val, oldVal){
            this.loadSelf()
        },
    },
})


// merge rack info with their rack units
var makeLumps = function(racks, units) {

    // for faster lookups
    var lookup = {}
    var byRID = {}
    for (var k=0; k<racks.length; k++) {
        var rack = racks[k]
        lookup[rack.RID] = rack

        // pre-populate empty rack
        var size = rack.RUs;
        var these = [];
        while(size--) these.push({
            DID:0, 
            RID:0, 
            Hostname:'', 
            Mgmt:'', 
            IPs:'', 
            RU: size+1, 
            Height: 1,
            badHeight: false,
            badHostname: false,
            badInternal: false,
            badMgmt: false,
            badIP: false,
        })
        byRID[rack.RID] = these
    }

    for (var i=0; i<units.length; i++) {
        var unit = units[i];
        /*
        if (unit.RID == 42 && unit.RU == 44) {
            console.log('break here')
        }
        */
        unit.newHostname = unit.Hostname
        unit.newHeight = unit.Height
        unit.newMgmt = unit.Mgmt
        unit.newIP = unit.IPs
        unit.badHeight = false
        unit.badHostname = false
        unit.badInternal = false
        unit.badMgmt = false
        unit.badIP = false
        var rack = lookup[unit.RID];
        if (rack) {
            byRID[unit.RID][rack.RUs - unit.RU] = unit
        }
    }

    var lumps = []
    for (var i=0; i<racks.length; i++) {
        var rack = racks[i];
        if (! rack || rack.RID == 0) continue
        var these = byRID[rack.RID]

        // for units greater than 1 RU, hide the slots consumed above
        // work our way up from the bottom
        for (var k=these.length - 1; k >= 0; k--) {
            for (var j=these[k].Height; j > 1; j--) {
                var x = k-j+1
                if (!these[x]) { continue }
                these[x].Height = 0;
            }
            these[k]['pingMgmt'] = ''
            these[k]['pingIP'] = ''
        }
        /*
        var above = 0;
        for (var k=0; k < these.length; k++) {
            these[k].above = above
            above++
            if (these[k].Hostname.length > 0) {
                above = 0;
            }
        }
        */
        lumps.push({rack: rack, units: these})
    }
    return lumps
}


/*
var rackViewVue = {
    methods: {
        rfilter: function(a, b, c) {
            if (this.RID == 0) {
                return a
            }
            if (this.RID == this.rack.RID) {
                    return a
            }
        },
    }
}
*/

// TODO: why you no like rackViewVue mixin?

var rackView = Vue.component('rack-view', {
    template: '#tmpl-rack-view',
    props: ['layouts', 'RID', 'audit'],
    //mixins: ['rackViewVue'],
    methods: {
        rfilter: function(a, b, c) {
            if (this.RID == 0) {
                return a
            }
            if (this.RID == this.rack.RID) {
                    return a
            }
        },
        move: function(ev) {
            console.log('move event:', ev)
        },
        moveUp: function(lay) {
            var did = lay.DID;
            var rid = lay.RID;
            var ru = lay.RU;
            var self = this;
            console.log('move up:', did, 'from:',ru);
            var off = self.layouts.rack.RUs - ru;
            var redux  = self.layouts.units[off-1];
            var device = self.layouts.units[off];

            ru++;
            var url = '/dcman/api/device/adjust/' + did;
            var adjust = {DID: device.DID, RID: device.RID, RU: ru, Height: device.Height};

            var moved = function(adjusted) {
                console.log('re adjusted:', adjusted);
                if (adjust.RU != adjusted.RU) {
                    console.log('fix me')
                } else {
                    redux.newHeight    = device.Height
                    redux.newHostname  = device.Hostname
                    redux.Height    = device.Height
                    redux.Hostname  = device.Hostname
                    redux.DID       = device.DID
                    redux.RID       = device.RID
                    redux.Mgmt      = device.Mgmt
                    redux.IPs       = device.IPs

                    device.newHeight    = 1
                    device.newHostname  = ''
                    device.Height    = 1
                    device.Hostname  = ''
                    device.DID       = 0
                    device.RID       = 0
                    device.Mgmt      = ''
                    device.IPs       = ''
                }

            }
            posty(url, adjust, 'PUT').then(moved);
        },
        moveDown: function(lay) {
            console.log('move down:', lay)
            var did = lay.DID;
            var rid = lay.RID;
            var ru = lay.RU;
            var self = this;
            console.log('move down:', did, 'from:',ru);
            var off = self.layouts.rack.RUs - ru;
            var redux  = self.layouts.units[off+1];
            var device = self.layouts.units[off];

            ru--;
            var url = '/dcman/api/device/adjust/' + did;
            var adjust = {DID: device.DID, RID: device.RID, RU: ru, Height: device.Height};

            var moved = function(adjusted) {
                console.log('re adjusted:', adjusted);
                if (adjust.RU != adjusted.RU) {
                    console.log('fix me')
                } else {
                    redux.newHeight    = device.Height
                    redux.newHostname  = device.Hostname
                    redux.Height    = device.Height
                    redux.Hostname  = device.Hostname
                    redux.DID       = device.DID
                    redux.RID       = device.RID
                    redux.Mgmt      = device.Mgmt
                    redux.IPs       = device.IPs

                    device.newHeight    = 1
                    device.newHostname  = ''
                    device.Height    = 1
                    device.Hostname  = ''
                    device.DID       = 0
                    device.RID       = 0
                    device.Mgmt      = ''
                    device.IPs       = ''
                }

            }
            posty(url, adjust, 'PUT').then(moved);
        },
        rackheight: function(lay) {
            return 'rackheight' + lay.Height;
        },
        // TODO make common, pass in field of interest
        changeIP: function(lay) {
            if (! validIP(lay.newIP.trim())) {
                lay.badIP = true;
                return
            }
            var url = ifaceViewURL;
            url += '?did=' + lay.DID + '&ipv4=' + lay.IPs;
            get(url).then(function(data) {
                // TODO add error handling
                var ipinfo = data[0]
                console.log("IPINFO:", ipinfo);
                var ip = {IID: ipinfo.IID, IPv4: lay.newIP}
                posty(ipURL + ipinfo.IID, ip, 'PATCH').then(function(updated) {
                    console.log("UPDATED:",updated)
                    lay.badIP = false;
                    lay.IP = lay.newIP;
                }) 
            })

        },
        changeMgmt: function(lay) {
            if (! validIP(lay.newMgmt.trim())) {
                lay.badMgmt = true;
                return
            }
            var url = ifaceViewURL;
            url += '?did=' + lay.DID + '&ipv4=' + lay.Mgmt;
            get(url).then(function(data) {
                // TODO add error handling
                var ipinfo = data[0]
                console.log("IPINFO:", ipinfo);
                var ip = {IID: ipinfo.IID, IPv4: lay.newMgmt}
                posty(ipURL + ipinfo.IID, ip, 'PATCH').then(function(updated) {
                    console.log("UPDATED:",updated)
                    lay.badMgmt = false;
                    lay.Mgmt = lay.newMgmt;
                }) 
            })

        },
        rename: function(lay) {
            if (lay.newHostname.trim().length === 0) {
                lay.badHostname = true
                return
            }
            if (lay.newHostname.trim() === lay.Hostname.trim()) {
                lay.badHostname = false
                return
            }
            // verify that new hostname doesn't already exist
            getDevice(lay.newHostname,'?hostname=').then(function(device) {
                if (! device) {
                    var newname = {DID: lay.DID, Hostname: lay.newHostname}
                    posty(deviceURL + lay.DID, newname, 'PATCH').then(function(good) {
                        lay.badHostname = false;
                    }).catch(function(fail) {
                        console.log('rename patch fail:', fail)
                    })
                } else {
                    lay.badHostname = true;
                }
            }).catch(function(fail) {
                    console.log('rename fail:', fail)
            })
        },
        resize: function(lay) {
            if (lay.newHeight < 1) {
                lay.badHeight = true;
                return
            }
            var self = this;
            var url = '/dcman/api/device/adjust/' + lay.DID;
            var newsize = {DID: lay.DID, RID: lay.RID, RU: lay.RU, Height: lay.newHeight}
            var resized = function(adjusted) {
                if (adjusted.Height == lay.Height) {
                    lay.badHeight = true;
                    return
                }
                if (lay.newHeight > lay.Height) {
                    var to = self.layouts.rack.RUs - lay.RU;
                    var from = to - lay.newHeight + 1;
                    for (var i=from; i < to; i++ ) {
                        self.layouts.units[i].Height = 0;
                    }
                } else if (lay.newHeight < lay.Height) {
                    var from = self.layouts.rack.RUs - lay.RU - lay.Height;
                    var to   = from + (lay.Height - lay.newHeight) + 1
                    for (var i=from; i < to; i++ ) {
                        self.layouts.units[i].Height = 1;
                    }
                }
                lay.Height = lay.newHeight
                lay.badHeight = false;
            }
            posty(url, newsize, 'PUT').then(resized);
        },
        okUp: function(lay) {
            if (! lay) return false;
            var ru = lay.RU
            var off = this.layouts.rack.RUs - ru;
            if (off < 1) return false;
            var device = this.layouts.units[off];
            if (! device || device.Hostname.length < 3) return false; // ignore empty units

            var space = ru + device.Height
            for (var i=device.RU +1 ; i <= space; i++) {
                var inv = this.layouts.rack.RUs - i;
                var unit = this.layouts.units[inv];
                if (unit && unit.Hostname && unit.Hostname.length > 3) {
                    return false
                }
            }
            return true;
        },
        okDown: function(lay) {
            if (! lay) return false;
            var ru = lay.RU;
            if (ru === 1) return false;
            var off = this.layouts.rack.RUs - ru;
            var device = this.layouts.units[off];
            if (! device || device.Hostname.length < 3) return false; // ignore empty units
            if (device.Hostname === 'hyp22020') {
                console.log('test device:',device)
            }

            for (var i=ru-1 ; i > 0; i--) {
                var inv = this.layouts.rack.RUs - i;
                var unit = this.layouts.units[inv];
                if (unit && unit.Hostname && unit.Hostname.length > 3) {
                    //console log("RACK:", this.layouts.rack.Label, "RU:", unit.RU, "HT:", unit.Height, "ru:", ru)
                    return (unit.RU + unit.Height) < ru;
                    //continue
                }
            }
            return true
        }
    }
})

//
// RACK LAYOUT
//

var rackLayoutVue = {
    data: function() {
        return {
            dataURL: deviceListURL,
            STI: 1,
            RID: 0,
            sites: [], 
            racks: [],
            site: '',
            audit: false,
            lumpy:[],
        }
    },

    created: function () {
        this.loadSelf()
    },
    route: { 
          data: function (transition) {
            var self = this;
            return Promise.all([
                getSiteLIST(), 
           ]).then(function (data) {
                return {
                    sites: data[0],
                }
           })
        }
    },
    methods: {
        rfilter: function(a, b, c) {
            if (this.RID == 0) {
                return a
            }
            if (this.RID == a.rack.RID) {
                    return a
            }
        },
        loadSelf: function () {
             var self = this

             var url = this.dataURL;

             if (self.RID > 0) {
                 url += "?rid=" + self.RID
             } else if (self.DID > 0) {
                 url += "?sti=" + self.STI
             }

             fetchData(url, function(units) {
                 url = rackURL + "?STI=" + self.STI;

                 fetchData(url, function(racks) {
                     if (racks) {
                         racks.unshift({RID:0, Label:''})
                         self.racks = racks
                         self.lumpy = makeLumps(racks, units)
                     }
                 })
             })
        },
        ping: function() {
            var url = "http://10.100.182.16:8080/dcman/api/pings?debug=true";
            var ips = [];
            for (var i=0; i < this.lumpy.length; i++) {
                var lump = this.lumpy[i];
                if (this.RID > 0 && lump.rack.RID != this.RID) continue;
                for (var k=0; k < lump.units.length; k++) {
                    var x = lump.units[k];
                    if (validIP(x.Mgmt)) ips.push(x.Mgmt);
                    if (validIP(x.IPs)) ips.push(x.IPs);
                }
            }
            var list = ips.join(",");
            var self = this
            postForm(url, {iplist: list}, function(xhr) {
               if (xhr.readyState == 4 && xhr.status == 200) {
                   var pinged = JSON.parse(xhr.responseText)
                    for (var i=0; i < self.lumpy.length; i++) {
                        for (var k=0; k < self.lumpy[i].units.length; k++) {
                            var unit = self.lumpy[i].units[k]
                            if (unit.Mgmt && unit.Mgmt.length > 0) { 
                                self.lumpy[i].units[k].pingMgmt = pinged[unit.Mgmt] ? "ok" : "-"
                            }
                            if (unit.IPs && unit.IPs.length > 0) 
                                self.lumpy[i].units[k].pingIP = pinged[unit.IPs] ? "ok" : "-"
                        }
                    }
               }
             });
        }
  },
  watch: {
    'STI': function(val, oldVal) {
            this.RID = 0
            this.loadSelf()
        },
    'RID': function(val, oldVal) {
            console.log('RID is now:', val)
        },
    },
}

var rackLayout = Vue.component('rack-layout', {
    template: '#tmpl-rack-layout',
    mixins: [rackLayoutVue],
})

var loginVue = {
    data: function() {
        return {
            username: '',
            password: '',
            placeholder: 'first.last@pubmatic.com',
            errorMsg: ''
        }
    },
    methods: {
        cancel: function() {
            router.go('/')
        },
        login: function(ev) {
            var url = '/dcman/api/login';
            var data = {Username: this.username, Password: this.password};
            var self = this;
            var results = function(xhr) {;
                if (xhr.readyState == 4) {
                    if (xhr.status == 200) {
                        var user = JSON.parse(xhr.responseText)
                        self.$dispatch('user-auth', user)
                        router.go('/')
                        return
                    }
                    console.log('login resp:' + xhr.responseText)
                    if (xhr.responseText.length > 0) {
                        var msg = JSON.parse(xhr.responseText)
                        self.errorMsg = msg.Error
                    }
                }
            };
            postIt(url, data, results)
            ev.preventDefault()
        },
    },
}


var mLogin = Vue.component('user-login', {
    template: '#tmpl-user-login',
    mixins: [loginVue],
})


var logoutVue = {
    methods: {
        cancel: function() {
            router.go('/')
        },
        logout: function(ev) {
            this.$dispatch('logged-out')
            router.go('/')
        },
    }
}

var mLogout = Vue.component('user-logout', {
    template: '#tmpl-user-logout',
    mixins: [logoutVue],
})


// grid component with paging and sorting
var pagedGrid = Vue.component('paged-grid', {
    template: '#tmpl-paged-grid',
    props: {
        data: Array,
        columns: Array,
        linkable: Function,
        linkpath: Function,
        //movePages: Function,
        startRow: Number,
        rowsPerPage: Number
    },
    data: function() {
        var sortOrders = {}
        if (this.columns) {
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
        }
        return {
              sortKey: '',
              sortOrders: sortOrders,
        }
    },
    computed: {
        rowStatus: function() {
            if (! this.rowsPerPage) {
                return this.data.length + ((this.data.length === 1) ? ' row' : ' rows')
            }
            return (
                (this.startRow / this.rowsPerPage + 1) +
                ' out of ' +
                (Math.ceil(this.data.length / this.rowsPerPage))
            )
        }
    },
    methods: {
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
        movePages: function(amount) {
            var newStartRow = this.startRow + (amount * this.rowsPerPage);
            if (newStartRow >= 0 && newStartRow < this.data.length) {
                this.startRow = newStartRow;
            }
        },
/*
        resetStartRow: function() {
            this.startRow = 0;
        },
*/
    }
});


var homePageVue = {
    data: function() {
        return {
            title: "PubMatic Datacenters",
            rows: [],
            columns: [ "Site", "Servers", "VMs" ],
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            var self = this;
            var url = '/dcman/api/summary/';
            fetchData(url, function(data) {
                self.rows = data
            })
        },
    },
}

var tallyMIX = {
    methods: {
        rowid: function(entry) {
            return "sti-" + entry.STI
        },
        linkable: function(key) {
            return (key == 'Label')
        },
        linkpath: function(entry, key) {
            if (key == 'Label') return '/rack/edit/' + entry['RID']
        }
    },
}

var tallyho = childTable("tally-table", "#tmpl-base-table", [tallyMIX])

var mHome = Vue.component('home-page', {
    template: '#tmpl-home-page',
    mixins: [homePageVue],
})


// Assign the new router
//var router = new VueRouter({history: true})
var router = new VueRouter()

// Assign your routes
router.map({
    '/auth/login': {
        component: Vue.component('user-login')
    },
    '/auth/logout': {
        component:  Vue.component('user-logout')
    },
    '/admin/tags': {
        component: Vue.component('tag-edit')
    },
    '/ip/list': {
        component:  Vue.component('ip-list')
    },
    '/ip/reserve': {
        component:  Vue.component('ip-reserve')
    },
    '/ip/types': {
        component:  Vue.component('ip-types')
    },
    '/ip/type/edit/:IPT': {
        component:  Vue.component('iptype-edit')
    },
    '/vlan/edit/:VLI': {
        component:  Vue.component('vlan-edit')
    },
    '/vlan/list': {
        component:  Vue.component('vlan-list')
    },
    '/device/add/:STI/:RID/:RU': {
        component: Vue.component('device-add'),
        name: 'device-add'
    },
    '/device/edit/:DID': {
        component: Vue.component('device-edit')
    },
    '/device/list': {
        component:  Vue.component('device-list')
    },
    '/vm/edit/:VMI': {
        component: Vue.component('vm-edit')
    },
    '/vm/list': {
        component:  Vue.component('vm-list')
    },
    '/mfgr/edit/:MID': {
        component: Vue.component('mfgr-edit')
    },
    '/mfgr/list': {
        component:  Vue.component('mfgr-list')
    },
    '/part/add': {
        component:  Vue.component('part-edit')
    },
    '/part/edit/:PID': {
        component:  Vue.component('part-edit')
    },
    '/part/list/:STI': {
        component:  Vue.component('part-list')
    },
    '/part/load': {
        component:  Vue.component('part-load')
    },
    '/part/types': {
        component:  Vue.component('part-types')
    },
    '/part/use/:STI/:KID': {
        component:  Vue.component('part-use')
    },
    '/part/inventory': {
        component:  Vue.component('part-inventory')
    },
    '/rack/edit/:RID': {
        component:  Vue.component('rack-edit')
    },
    '/rack/list': {
        component:  Vue.component('rack-list')
    },
    '/rack/layout': {
        component:  Vue.component('rack-layout')
    },
    '/rma/create/:PID': {
        component:  Vue.component('rma-create')
    },
    '/rma/edit/:RMD': {
        component:  Vue.component('rma-edit')
    },
    '/rma/list': {
        component:  Vue.component('rma-list')
    },
    '/user/edit/:USR': {
        component:  Vue.component('user-edit')
    },
    '/user/list': {
        component:  Vue.component('user-list')
    },
    '/vendor/edit/:VID': {
        component:  Vue.component('vendor-edit')
    },
    '/vendor/list': {
        component:  Vue.component('vendor-list')
    },
    '/search/:searchText': {
        component:  Vue.component('search-for'),
        name: 'search'
    },
    '/': {
        component:  Vue.component('home-page')
    },
})

router.beforeEach(function (transition) {
    if (window.user_apikey.length == 0 && transition.to.path !== '/auth/login') {
        router.go('/auth/login')
        //transition.abort()
    } else {
        transition.next()
    }
})



    router.start(App, '#app')
/*
var baseUrl = 'https://pubmatic.okta.com/'
var oktaSignIn = new OktaSignIn({baseUrl: baseUrl});

oktaSignIn.renderEl(
  { el: '#okta-login-container' },
  function (res) {
    if (res.status === 'SUCCESS') {
      console.log('User %s successfully authenticated %o', res.user.profile.login, res.user);
      res.session.setCookieAndRedirect('https://example.com/');
    }
  }
);
*/

