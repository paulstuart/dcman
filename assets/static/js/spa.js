'use strict';

var pingURL = "http://10.100.182.16:8080/dcman/api/pings?debug=true";

var userInfo = {};

var mySTI = 1;

var rackData = {
    STI: 0,
    list: [],
}

var deviceListURL = '/dcman/api/device/ips/'
var deviceTypesURL = '/dcman/api/device/type/'

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

function getTagList() {
  return get(tagURL).then(function(result) {
       console.log('tag list fetched:', result.length);
       return result;
    })
    .catch(function(x) {
      console.log('Could not tag list: ', x);
    });
}

function getPartTypes() {
  return get(partTypesURL).then(function(result) {
       console.log('part types fetched:', result.length);
       return result;
    })
    .catch(function(x) {
      console.log('Could not load part types: ', x);
    });
}

function getDeviceTypes() {
  return get(deviceTypesURL).then(function(result) {
       console.log('device types fetched:', result.length);
       return result;
    })
    .catch(function(x) {
      console.log('Could not load device types: ', x);
    });
}

function getIPTypes() {
  return get(iptypesURL).then(function(result) {
       console.log('ip types fetched:', result.length);
       return result;
    })
    .catch(function(x) {
      console.log('Could not load ip types: ', x);
    });
}

function getDeviceLIST(STI) {
  var url = deviceListURL;
  if (STI > 0) {
    url += "?STI=" + STI
  }
  return get(url).then(function(result) {
       console.log('device list fetched:', result.length);
       return result;
    })
    .catch(function(x) {
      console.log('Could not load device list: ', x);
    });
}


function getDevice(DID) {
  // Use the fetch API to get the information
  // fetch returns a promise
  var url = deviceURL;
  if (DID && DID > 0) {
    //url += "?DID=" + DID
    url += DID
  }
  return get(url).then(function(result) {
       console.log('device fetched')
       return result;
    })
    .catch(function(x) {
      //throw new Error('Could not load serverlist: ');
      console.log('Could not load device: ', x);
    });
}

function getInterfaces(device) {
  //var device = data[0];
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
    //console.log('PORTS ***********', ports, 'LEN:', ports.length)
    device.ips = ips
    device.interfaces = good
    //self.netRows = uniqueInterfaces(data)
    //console.log('net row cnt:', self.interfaces.length)
    return device
   })
}

// device info with associated interface / IPs
var completeDevice = function(DID) {
   return getDevice(DID).then(getInterfaces); 
}

function fetchRacks(STI) {
    var url = rackURL;
    if (STI > 0) {
        url += "?STI=" + STI
    }
     return get(url).then(function(result) {
           console.log('rack fetched:', result.length);
             result.unshift({RID:0, Label:''})
           rackData.list = result;
           return result;
        })
        .catch(function(x) {
          //throw new Error('Could not load serverlist: ');
          console.log('Could not load racks: ', x);
        });
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
    var octs = ip.split('.')
    if (octs.length != 4) return false
    for (var i=0; i<4; i++) {
        if (octs[i].length == 0) return false
        var val=parseInt(octs[i])
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

// common stuff for edits 
var editVue = {
    created: function () {
        this.loadSelf()
    },
    methods: {
        saveSelf: function() {
            if (this.preSave) {
                this.preSave
            }
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



var ipgridMIX = {
    props: [ 'STI', 'Host', 'Type'],
    created: function(ev) {
        console.log('t estmix created!');
    },
    ready: function(ev) {
        console.log('t estmix ready!');
    },
    methods: {
        subFilter: function(a, b, c) {
            if (! this.Type && ! this.Host) {
                return a
            }
            if (this.Type == a.Type && ! this.Host) {
                return a
            }
            if (this.Host == a.Host && ! this.Type) {
                return a
            }
            if (this.Host == a.Host && this.Type == a.Type) {
                return a
            }
        },
        linkable: function(key) {
            return (key == 'Hostname')
        },
        linkpath: function(entry, key) {
            if (entry.Host != 'VM') {
                return '/device/edit/' + entry['ID']
            }
            return '/vm/edit/' + entry['ID']
        }
    },
}

var ippageMIX = {
    created: function(ev) {
        console.log('ip pageMIX created!');
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
            Type: '',
            Host: '',
            site: 'blah',
            STI: 1,
            searchQuery: '',
            sites: [],
                // TODO: pull list from DB
            typelist: [
               '',
                'IPMI',
                'Internal',
                'Public',
                'VIP',
            ],
            // TODO: kindlist should be populated from device_types
            hostlist: [
               '',
                'VM',
                'Server',
                'Switch',
            ],
        }
    },
    ready: function() {
        console.log('created ips:', this.rows.length)
        for (var i=0; i<this.rows.length; i++) {
            var ip = this.rows[i];
            //console.log('my IP:', ip.IP, 'What:', ip.What, 'Kind:', ip.Kind)
                console.log('my IP:', ip.IP, 'What:', ip.What, 'Kind:', ip.Kind)
                continue
            if (ip.What === 'server' && ip.Kind === 'internal') 
                console.log('my IP:', ip.IP, 'What:', ip.What, 'Kind:', ip.Kind)
        }
    },
    route: { 
          data: function (transition) {
            var self = this;
            console.log('server list promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(), 
           ]).then(function (data) {
              console.log('server list promises returning. site label:', self.site, 'STI:', self.STI)
              return {
                sites: data[0],
                //site: getSiteName(data[0], self.STI),
              }
            })
          }
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
}


var dg = childTable("ip-grid", "#tmpl-base-table", [ipgridMIX])

var Netlist = Vue.component('ip-list', {
    template: '#tmpl-ip-list',
    mixins: [ipload, ippageMIX, commonListMIX],
})


//
// User List
//

var userMIX = {
    methods: {
        linkable: function(key) {
            return (key == 'Login')
        },
        linkpath: function(entry, key) {
            return '/user/edit/' + entry['ID']
        }
    },
}

var ug = childTable("user-grid", "#tmpl-base-table", [userMIX])

var userListVue = {
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
    }
}

var UserList = Vue.component('user-list', {
    template: '#tmpl-user-list',
    mixins: [userListVue],
})

//
// USER EDIT
//

var userEditVue = {
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
            return this.User.ID
        },
        myself: function() {
            return this.User
        },
        loadSelf: function () {
            var self = this;
            var id = this.$route.params.UID;
            if (id > 0) {
                var url = this.dataURL + id;

                fetchData(url, function(data) {
                    self.User.Load(data);
                })
            }
        },
    },
}

var uEdit = Vue.component('user-edit', {
    template: '#tmpl-user-edit',
    mixins: [userEditVue, editVue],
})



//
// Server Add
//
/*
var serverAddVue = {
    data: function() {
        return {
            sites: [],
            racks: [],
            tags: [],
            Description: '',
            Server: new(Server),
        }
    },
    created: function () {
        this.Server.STI    = 0;
        this.Server.Height = 1;
        this.Server.ID     = 0;
        this.Server.TID    = 0;
        this.Server.Rack   = 0;
        this.Server.RID    = parseInt(this.$route.params.RID);
        this.Server.RU     = parseInt(this.$route.params.RU);
        this.loadTags()
        this.loadRacks()
        this.sites = getSizeList()
    },
    methods: {
        saveSelf: function(event) {
            console.log('send update event: ' + event);
            postIt(serverURL + "?debug=true", this.Server, this.showList)
        },
        Deleted: function() {
            // stub necessary to use server-edit 
        },
        showList: function(ev) {
            this.$route.router.go(window.history.back())
        },
        loadRacks: function () {
            var self = this;
            var url = rackURL;
            if (this.Server.STI > 0) {
                url += "?STI=" + self.Server.STI
            }
            fetchData(url, function(data) {
                if (self.Server.STI > 0) {
                    self.racks = data
                    return
                }
                // find our site
                for (var i=0; i<data.length; i++) {
                    if (data[i].RID == self.Server.RID) {
                        var sti = data[i].STI;
                        // now delete all racks not with this site
                        var keep = [];
                        for (var k=0; k<data.length; k++) {
                            if (data[k].STI == sti) {
                                keep.push(data[k])
                            }
                        }
                        self.Server.STI = sti
                        self.racks = keep 
                        break
                    }
                }
            })
        },
        loadTags: function () {
            var self = this;
            fetchData(tagURL, function(data) {
                self.tags = data
            })
        },
        getMacAddr: function(ev) {
            //var url = '/dcman/data/server/discover/' + this.Server.IPIpmi;
            var url = 'http://10.100.182.16:8080/dcman/data/server/discover/' + this.Server.IPIpmi;
            var self = this
            fetchData(url, function(data) {
                self.Server.MacPort0 = data.MacEth0
                console.log("MAC DATA:", data)
             })
             ev.preventDefault();
            return false;
        },

    },

    watch: {
        "Server.STI": function(older, newer) {
            this.loadRacks()
        },
        "Server.ID": function(older, newer) {
            console.log("new server id:", this.Server.ID)
        },
    },
    events: {
        'server-reload': function(x) {
            var id = this.$route.params.SID;
            console.log('* reload server ID:', id, 'current ID:', this.Server.ID, 'SID:', this.SID)
            if (id != this.Server.ID) {
                console.log('********************** NEW server ID:', id)
                this.loadServer()
            }
        }
    }
}
*/

var ipURL = '/dcman/api/network/ip/'
var deviceURL = '/dcman/api/device/view/'
var iptypesURL = '/dcman/api/network/ip/type/'
var ifaceURL = '/dcman/api/interface/'
var ifaceViewURL = '/dcman/api/interface/view/'

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
            /*
            postIt(ipURL + iid, null, null, 'DELETE')
            return false
            */
            deleteIt(ipURL + iid, function(xhr) {
                //console.log('del state:', xhr.readyState, 'status:', xhr.status)                
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

var netgrid = childTable("network-grid", "#tmpl-network-grid", [ipMIX])//, [noLinks])

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
            //postIt(ifaceURL + ifd, null, null, 'DELETE')
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
            //ev.preventDefault()
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
            racks: [],
            device_types: [],
            //ports: [],
            tags: [],
            ipTypes: [],
            ipRows: [],
            interfaces: [],
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
            console.log('ROUTE TRANS:',transition)
            console.log('route did:',this.$route.params.DID)
            return Promise.all([
                getSiteLIST(false), 
                getDeviceTypes(), 
                getTagList(),
                getIPTypes(),
                completeDevice(this.$route.params.DID), 
                //fetchRacks(this.STI), 
           ]).then(function (data) {
               // console.log('NEW DEVICE:', data[3]);
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
/*
    created: function () {
        this.loadTags()
        //this.loadSelf()
    },
*/
    methods: {
        saveSelf: function(event) {
            if (this.Device.DID == 0) {
                console.log('save new device');
                postIt(deviceURL + "?debug=true", this.Device, this.showList)
                return
            }
            console.log('update device id: ' + this.Device.DID);
            postIt(deviceURL + this.Device.DID + "?debug=true", this.Device, this.showList, 'PATCH')
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            deleteIt(deviceURL + this.Device.DID, this.showList)
        },
        showList: function(ev) {
            this.$route.router.go(window.history.back())
        },
        loadRacks: function () {
             var self = this;
            console.log("RACK URL:", rackURL + "?STI=" + self.Device.STI)
             fetchData(rackURL + "?STI=" + self.Device.STI, function(data) {
                 self.racks = data
             })
        },
        loadDevice: function () {
             var self = this;

             var id = this.$route.params.DID;
             //console.log('loading device ID:', id)
             if (id > 0) {
                 var url = deviceURL + id;

                 fetchData(url, function(data) {
                     self.Device.Load(data);
                     self.loadRacks()

                     var url = ifaceViewURL + '?DID=' + id
                     fetchData(url, function(data) {
                        if (! data) return
                        
                         var ips = []
                         var ports = {}
                         for (var i=0; i<data.length; i++) {
                             var ip = data[i]
                             if (! (ip.IFD in ports)) {
                                ports[ip.IFD] = ip
                            }
                            if (ip.IP) ips.push(ip)
                        }
                        var good = [];
                        for (var ifd in ports) {
                            good.push(ports[ifd])
                        }
                        //console.log('PORTS ***********', ports, 'LEN:', ports.length)
                        self.ipRows = ips
                        self.interfaces = good
                        //self.netRows = uniqueInterfaces(data)
                        console.log('net row cnt:', self.interfaces.length)
                     })
                 })
               }
               fetchData(iptypesURL, function(data) {
                 self.ipTypes = data
               })
        },
        loadSelf: function () {
             var self = this;

             var id = this.$route.params.DID;
             //console.log('loading device ID:', id)
             if (id > 0) {
                 var url = deviceURL + id;

                 fetchData(url, function(data) {
                     self.Device.Load(data);
                     //self.loadRacks()

                     var url = ifaceViewURL + '?DID=' + id
                     fetchData(url, function(data) {
                        if (! data) return
                        
                         var ips = []
                         var ports = {}
                         for (var i=0; i<data.length; i++) {
                             var ip = data[i]
                             if (! (ip.IFD in ports)) {
                                ports[ip.IFD] = ip
                            }
                            if (ip.IP) ips.push(ip)
                        }
                        var good = [];
                        for (var ifd in ports) {
                            good.push(ports[ifd])
                        }
                        //console.log('PORTS ***********', ports, 'LEN:', ports.length)
                        self.ipRows = ips
                        self.interfaces = good
                        //self.netRows = uniqueInterfaces(data)
                        console.log('net row cnt:', self.interfaces.length)
                     })
                 })
               }
               fetchData(iptypesURL, function(data) {
                 self.ipTypes = data
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

    watch: {
        "Device.STI": function(older, newer) {
            console.log("new device STI:", this.Device.STI)
            this.loadRacks()
        },
        "Device.DID": function(older, newer) {
            console.log("new device id:", this.Device.DID)
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
    attached: function () {
        console.log('ATTACHED VMI:', this.VMI)
        //this.loadSelf()
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
            //postIt(ipURL + iid, null, null, 'DELETE')
            deleteIt(ipURL + iid, function(xhr) {
                //console.log('del state:', xhr.readyState, 'status:', xhr.status)                
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    self.rows.splice(i, 1)
                }
            })
            return false
        },
        addIP: function() {
            var self = this;
            var data = {VMI: this.VMI, IPT: this.newIPT, IPv4: this.newIP}
            console.log("we will add IP info:", data)
            postIt(ipURL + '?debug=true', data, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    self.rows.push(data)
                    self.newIP = ''
                    self.newIPT = 0
                }
            })
            return false
        }
    }
}

//var netgrid = childTable("vm-ips", "#tmpl-vm-ips", [vmIpMIX])
//var netgrid = vue.Component("vm-ips", "#tmpl-vm-ips", [vmIpMIX])
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
    created: function () {
        this.loadSelf()
    },
    attached: function() {
         console.log("VM ATTACHED:", this.$route.params.VMI)
         var id = this.$route.params.VMI;
         if (id && id != this.VM.VMI) {
             this.loadSelf()
         }
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
        loadSelf: function () {
             var self = this;
             var id = this.$route.params.VMI;
             console.log('loading vm ID:', id)
             if (id > 0) {
                 this.ID = id
                 var url = this.url + id;

                 fetchData(url, function(data) {
                     self.VM.Load(data);
                 })
               }
        },
    },
    events: {
        'vm-reload': function(x) {
            var id = this.$route.params.VMI;
            if (id != this.VM.VMI) {
                console.log('********************** NEW server ID:', id)
                this.loadVM()
            }
        }
    }
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
/*
        'search-for': function (msg) {
            console.log('*** search-for event:', msg)
            this.search(msg)
        },
*/
        'search-again': function(text) {
            // relay event from navbar search
            this.$broadcast('search-again', text)
        },
    },
})


var deviceMIX = {
    props: [ 'RID' ],
    methods: {
        subFilter: function(a, b, c) {
            if (! this.RID) {
                return a
            }
            if (this.RID == 0) {
                return a
            }
            if (this.RID == a.RID) {
                return a
            }
        },
        linkable: function(key) {
            return (key == 'Hostname')
        },
        linkpath: function(entry, key) {
            return '/device/edit/' + entry['DID']
        }
    },
}

var sg = childTable("device-grid", "#tmpl-base-table", [deviceMIX])

var deviceListVue = {
  data: function() {
      console.log('device data returning')
      return {
          STI: mySTI,
          RID: 0,
          sites: [],
          racks: [],
          site: 'blah',
          searchQuery: '',
          gridData: [],
          gridColumns: [
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
            ]
        }
  },
    route: { 
          data: function (transition) {
            //var userId = transition.to.params.userId
            var self = this;
            console.log('device list promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(), 
                getDeviceLIST(self.STI), 
                fetchRacks(self.STI), 
           ]).then(function (data) {
              console.log('device list promises returning')
             return {
                sites: data[0],
                gridData: data[1],
                racks: data[2],
                //site: getSiteName(data[0], self.STI),
              }
            })
          }
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
         self = this
         self.RID = 0

         var url = deviceListURL;
         if (self.STI > 0) {
             url += "?sti=" + self.STI
         }
         fetchData(url, function(data) {
             self.gridData = data
         })
         self.loadRacks()
    },
  },
  watch: {
    'STI': function(val, oldVal){
            mySTI = val
            this.loadSelf()
        },
    },
}

var deviceList = Vue.component('device-list', {
    template: '#tmpl-device-list',
    mixins: [deviceListVue, commonListMIX],
})
//
// VLAN List
//

var vlanMIX = {
    methods: {
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/vlan/edit/' + entry['ID']
        }
    },
}

var vlg = childTable("vlan-grid", "#tmpl-base-table", [tableTmpl, vlanMIX])

var vlanListVue = {
    data: function() {
        return {
            dataURL: '/dcman/api/vlan/view/',
            listURL: '/vlan/list',
            sites: [],
            searchQuery: '',
            rows: [],
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
    },
    methods: {
        loadSelf: function () {
            var self = this;
            var url = this.dataURL;
            fetchData(url, function(data) {
                self.rows = data
            })
        },
    },
}

var VLANList = Vue.component('vlan-list', {
    template: '#tmpl-vlan-list',
    mixins: [vlanListVue, commonListMIX],
})

//
// VLAN Edit
//
var vlanEditVue = {
    data: function() {
        return {
            sites: [],
            VLAN: new(VLAN),
            dataURL: '/dcman/api/vlan/view/'
        }
    },
    created: function () {
        this.loadSelf()
    },
      attached: function() {
          console.log("DEVICE ATTACHED:", this.$route.params.SID)
           var id = this.$route.params.SID;
          if (id && id != this.SID) {
              this.loadSelf()
          }
      },
      methods: {
/*
          Updated: function(event) {
              console.log('send update event: ' + event);
              this.Server.TID = parseInt(this.Server.TID)
              postIt(serverURL + this.SID + "?debug=true", this.Server, this.showList, 'PATCH')
          },
          Deleted: function(event) {
              console.log('delete event: ' + event)
              postIt(serverURL + this.SID, null, this.showList, 'DELETE')
          },
*/
          myID: function() {
              return this.VLAN.ID
          },
          myself: function() {
              return this.VLAN
          },
          showList: function(ev) {
              router.go('/vlan/list')
          },
          loadSelf: function () {
               var self = this;

               var id = this.$route.params.VLID;
               console.log('loading vlan ID:', id)
               if (id > 0) {
                   var url = vlanURL + id;

                   fetchData(url, function(data) {
                       self.VLAN.Load(data);
                   })
                 }
          },
/*
          deleteSelf: function(event) {
              console.log('delete tag evnt: ' + event)
              postIt(this.url + this.TID, null, this.showList, 'DELETE')
          },
          saveSelf: function(event) {
              this.tag.TID = parseInt(this.tag.TID)
              console.log('update tag event: ' + event);
              if (this.tag.TID > 0) {
                  // postIt = function(url, data, fn, method) {

                  postIt(this.url + this.tag.TID + "?debug=true", this.tag, this.showList, 'PATCH')
              } else {
                  postIt(this.url + "?debug=true", this.tag, this.showList)
              }
              this.loadSelf()
          },
*/
    },
}

var vlanEdit = Vue.component('vlan-edit', {
    template: '#tmpl-vlan-edit',
    mixins: [editVue, vlanEditVue, commonListMIX, siteMIX],
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
    /*
    created: function () {
        this.loadSelf()
    },
    */
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
        reserveIPs: function() {
            var from = ip32(this.From)
            var to = ip32(this.To)
            for (var i = from; i <= to; i++) {
            }
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

var vmMIX = {
    methods: {
        linkable: function(key) {
            return (key == 'Hostname' || key == 'Server')
        },
        linkpath: function(entry, key) {
            if (key == 'Server') return '/device/edit/' + entry['DID']
            if (key == 'Hostname') return '/vm/edit/' + entry['VMI']
        }
    },
}
var vg = childTable("vm-grid", "#tmpl-base-table", [vmMIX])

var vmListVue = {
  data: function() {
      return {
      STI: 1,
      sites: [],
      site: 'blah',
      searchQuery: '',
      gridData: [],
      gridColumns: [
           "Site",
           "Server",
           "Hostname",
               /*
           "Private",
           "Public",
           "VIP",
           */
           "Profile",
           "Note",
        ]
    }
  }
  ,

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
             self.gridData = data
         })
         self.loadRacks()
        
    },
  },

  watch: {
    'STI': function(val, oldVal){
            this.loadSelf()
        },
    },
}

var VMList = Vue.component('vm-list', {
    template: '#tmpl-vm-list',
    mixins: [vmListVue, siteMIX, commonListMIX],
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


var inventoryMIX = {
    methods: {
        myFilter: function(a, b, c) {
            return a
        },
        /*
        updated: function(ev) {
            var qty = ev.target.value
            var arr = ev.target.id.split("-")
            updateQty(arr[1], qty)
        },
        sortBy: function (key) {
          this.sortKey = key
          this.sortOrders[key] = this.sortOrders[key] * -1
        },
        delPart: function (ev) {
            var id = parseInt(ev.target.id.split('-')[1])
            postIt(partURL + id, null, function(xreq) {
                if (xreq.readyState == 4) {
                     if (xreq.status != 200) {
                        alert("Oops:" + xreq.responseText);
                        return
                     }
                    self.$dispatch('item-added', 'incoming!')
                }
            }, 'DELETE')
        },
        usePart: function(ev) {
            var id = parseInt(ev.target.id.split('-')[1]);
            this.$dispatch('select-part', id)
        },
        addPart: function(ev) {
            var id = parseInt(ev.target.id.split('-')[1]);
            this.$dispatch('new-item', id) 
        }, 
        */
        linkable: function(key) {
            return (key == 'Description')
        },
        linkpath: function(entry, key) {
            return '/part/inventory/' + entry['STI'] + '/' + entry['KID']
        },
        slotid: function(entry) {
            alert(entry)
        },
        rowid: function(entry) {
            return "myrow"
        },
    },
}

var dg = childTable("inventory-grid", "#tmpl-base-table", [inventoryMIX])

var inventoryVue = {
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
      gridColumns: ['Site', 'Description', 'PartNumber', 'Mfgr', 'Qty'],
    }
  },
  created: function () {
      this.loadSelf()
  },
  methods: {
    updated: function(event) {
        console.log('the event: ' + event)
    },
    loadSelf: function () {
         var self = this;
         var url = inURL;
         if (self.STI > 0) {
            url += "?sti=" + self.STI
         }
         fetchData(url, function(data) {
             self.partData = data
         })
     },
  },
  watch: {
    'STI': function(val, oldVal){
            this.loadSelf()
        },
    },
}


var iiPart = Vue.component('part-inventory', {
    template: '#tmpl-part-inventory',
    mixins: [inventoryVue, commonListMIX, siteMIX],
})


var partUseVue = {
  data: function() {
      return {
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
      findhost: function(ev) {
          var self = this;
          console.log("find hostname:",this.hostname);
          fetchData("/dcman/api/server/hostname/" + this.hostname, function(data, status) {
                var enable = (status == 200);
                buttonEnable(document.getElementById('use-btn'), enable)
                self.DID = enable ? data.ID : 0;
            })
      },
  },
}

var usePart = Vue.component('part-use', {
    template: '#tmpl-use-part',
    mixins: [partUseVue],
})

//
// Parts
//

var partListVue = {
    data: function() {
        return {
            showgrid: true,
            isgood: true,
            isbad: false,
            DID: 0,
            STI: 4,
            PID: 0,
            available: [],
            sites: [],
            hostname: '',
            other: '',
            searchQuery: '',
            partData: [],
            ktype: 1,
            kinds: [
                {id: 1, name: "All Parts"},
                {id: 2, name: "Good Parts"},
                {id: 3, name: "Bad Parts"},
            ],
            gridColumns: [
               "Site",
               //"Hostname",
               "Serial",
               "PartType",
               "PartNumber",
               "Description",
               "Mfgr",
               "Bad",
            ]
        }
    },
    created: function () {
        this.loadSelf()
    },
    route: { 
          data: function (transition) {
            //var userId = transition.to.params.userId
            return {
              sites: getSiteLIST(true), //sitePromise,
            }
          }
    },
    methods: {
        loadSelf: function () {
        var self = this
        var url = partURL;
        if (self.STI > 0) {
            url += "?sti=" + self.STI
        }
        fetchData(url, function(data) {
            if (data) {
                for (var i=0; i<data.length; i++) {
                    data[i].Serial = data[i].Serial || '';
                }
             }
             self.partData = data
        })
    },
    partOk: function(ev) {
          var part = {
            KID: parseInt(document.getElementById("new-kid").value),
            Serial: document.getElementById("add-sn").value,
            Unused: 1,
          }
          postIt(mypartURL, part)
      },
      useOk: function(ev) {
          var pid = document.getElementById("PID").value
          var part = {
            PID: parseInt(pid),
            STI: this.STI,
            DID: this.DID,
            Unused: 0,
          }
          postIt(mypartURL + pid, part, null, "PATCH")
      },
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
    },
    watch: {
        'STI': function() {
            this.loadSelf()
        }
    }
}


/*
var iList = Vue.component('part-inventory', {
    template: '#tmpl-part-inventory',
    mixins: [inventoryVue, siteMIX],
})
*/


var partsMIX = {
    methods: {
        subFilter: function(a, b, c) {
            if (this.kfilter == 1) {
                return a
            }
            if (this.kfilter == 2 && ! a.Bad) {
                return a
            }
            if (this.kfilter == 3 && a.Bad) {
                return a
            }
        },
        linkable: function(key) {
            return (key == 'Description')
        },
        linkpath: function(entry, key) {
            return '/part/edit/' + entry['PID']
        }
    }
}

var partt = Vue.component('parts-table', {
    template: '#tmpl-base-table',
    props: ['kfilter', 'filterKey', 'myfn'],
    mixins: [tableTmpl, partsMIX],
})

//
// All Parts
//
var pList = Vue.component('part-list', {
    template: '#tmpl-part-list',
    mixins: [partListVue, commonListMIX],
})

//
// USE PART
//

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
            return (key == 'Description')
        },
        linkpath: function(entry, key) {
            return '/rma/edit/' + entry['RMAID']
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
                "RMAID",
                "Description",
                "Hostname",
                "ServerSN",
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
//var vendg = childTable("vendor-grid", "#tmpl-base-table")

var vendorList = Vue.component('vendor-list', {
    template: '#tmpl-vendor-list',
    mixins: [vendorListVue],
})

var vendorEditVue = {
    data: function() {
        return {
            Vendor: new(Vendor),
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
// PART EDIT
//
var partEditVue = {
    data: function() {
        var part = new(Part);
        part.PID = 0;
        part.DID = 0;
        part.STI = 0;
        part.PTI = 0;
        part.Bad = false;
        part.Used = false;
       return {
            sites: [],
            types: [],
            STI: 4,
            Part: part,
       }
    },
    computed: {
        disableSave: function() {
            console.log("part PTI:", this.Part.PTI)
            if (this.Part.PTI == 0) {
                return (this.Part.STI == 0 || this.Part.PTI == 0 || this.Part.Description.length == 0)
                //return (this.Part.Description.length == 0)
            }
/*
            return false;
            return this.Part.Description.length == 0 || this.Part.Mfgr.length == 0
*/
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
            }
          }
    },
    methods: {
        showList: function(ev) {
            router.go('/part/list')
        },
        saveSelf: function(event) {
            console.log('send event: ' + event)
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
            ev.preventDefault();
            var data = {
                STI: this.STI,
                OldPID: this.Part.PID,
                Created: new Date(),
            };
            postIt(rmaURL + "?debug=true", data, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    var rma = JSON.parse(xhr.responseText)
                    router.go('/rma/edit/' + rma.RMAID)
                }
            })
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
        findhost: function(ev) {
              var self = this;
              console.log("find hostname:",this.Part.Hostname);
                fetchData("api/server/hostname/" + this.Part.Hostname, function(data, status) {
                    var enable = (status == 200);
                     // buttonEnable(document.getElementById('use-btn'), enable)
                    self.Part.DID = enable ? data.ID : 0;
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
var rmaEditVue = {
  data: function() {
      return {
        sites: [],
        racks: [],
        tags: [],
        dataURL: rmaviewURL,
        RMA: new(RMA),
    }
  },
  created: function () {
      this.loadRMA()
  },

  methods: {
      Updated: function(event) {
          console.log('send update event: ' + event);
          postIt(rmaviewURL + this.RMA.RMAID + "?debug=true", this.RMA, this.showList, 'PATCH')
      },
      // TODO: use editVue mixin
      myID: function() {
          return this.RMA.RMAID
      },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            postIt(this.dataURL + this.myID(), null, this.showList, 'DELETE')
        },
      showList: function() {
          //ev.preventDefault();
          router.go('/rma/list')
      },
      loadRMA: function () {
          var self = this;
          var id = this.$route.params.RMAID;
          if (id > 0) {
              this.DID = id
              var url = rmaviewURL + id;

              fetchData(url, function(data) {
                  console.log('self 2', self);
                  self.RMA.Load(data);
              })
          }
      },
      findhost: function(ev) {
          var self = this;
          console.log("find hostname:",this.RMA.Hostname);
            fetchData("api/server/hostname/" + this.RMA.Hostname, function(data, status) {
                var enable = (status == 200);
               // buttonEnable(document.getElementById('use-btn'), enable)
                self.RMA.DID = enable ? data.ID : 0;
            })
      },
  },
}

var rEdit = Vue.component('rma-edit', {
    template: '#tmpl-rma-edit',
    mixins: [rmaEditVue],
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
        /*
    attached: function() {
        console.log("RACK ATTACHED:", this.$route.params[this.id])
        var id = this.$route.params[this.id];
        if (id && id != this.myID()) {
              this.loadSelf()
        }
    },
    */
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
        preSave: function() {
            this.Rack.Rack = parseInt(this.Rack.Rack)
            this.Rack.RUs = parseInt(this.Rack.RUs)
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

var rackMIX = {
    methods: {
        linkable: function(key) {
            return (key == 'Label')
        },
        linkpath: function(entry, key) {
            if (key == 'Label') return '/rack/edit/' + entry['RID']
        }
    },
}
var rackg = childTable("rack-grid", "#tmpl-base-table", [rackMIX])

var rackListVue = {
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
  },

  watch: {
    'STI': function(val, oldVal){
            this.loadSelf()
        },
    },
}

var rackList = Vue.component('rack-list', {
    template: '#tmpl-rack-list',
    mixins: [rackListVue, siteMIX],
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
        while(size--) these.push({RU: size+1, Height: 1})
        byRID[rack.RID] = these
    }

    for (var i=0; i<units.length; i++) {
        var unit = units[i]
        var rack = lookup[unit.RID]
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
                these[x].Height = 0;
            }
        }
        lumps.push({rack: rack, units: these})
        //console.log('lumpy rack:', rack.RID, 'len:', these.length)
    }
    return lumps
}


var rackLayoutVue = {
    data: function() {
        return {
            dataURL: deviceListURL,
            STI: 1,
            RID: 0,
            sites: [], 
            racks: [],
            site: '',
            layouts: [],
            lumpy:[]
        }
    },

    created: function () {
        this.loadSelf()
    },
    route: { 
          data: function (transition) {
            var self = this;
            console.log('rack layout promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(), 
           ]).then(function (data) {
              console.log('rack layout promises returning')
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
                 self.layouts = units

                 url = rackURL + "?STI=" + self.STI;

                 fetchData(url, function(racks) {
                     if (racks) {
                         racks.unshift({RID:0, Label:''})
                         self.racks = racks
                         self.lumpy = makeLumps(racks, units)
                         console.log("LUMPY STI:", self.STI)
                     }
                 })
             })
        },
        ping: function() {
            var url = "http://10.100.182.16:8080/dcman/api/pings?debug=true";
            var ips = [];
            for (var i=0; i < this.layouts.length; i++) {
                var x = this.layouts[i];
                if (validIP(x.Mgmt)) ips.push(x.IPMI);
                if (validIP(x.IPs)) ips.push(x.IPs);
            }
            //postIt(url, {ips: ips}, function(data) {
            var list = ips.join(",");

            postForm(url, {iplist: list}, function(xhr) {
               if (xhr.readyState == 4 && xhr.status == 200) {
                   var pinged = JSON.parse(xhr.responseText)
                    for (ip in pinged) {
                        var cell = document.getElementById("ip-" + ip);
                        cell.innerHTML = (pinged[ip] ? "ok" : "-")
                        
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

// LAYOUT
var rackView = Vue.component('rack-view', {
    template: '#tmpl-rack-view',
    props: ['rack', 'layout', 'layouts', 'RID'],
    mixins: ['rackViewVue'],
})

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
            //router.go('/')
            ev.preventDefault()
        },
    },
}

            /*
login resp:{
  "ID": 1,
  "RealID": 0,
  "Login": "pstuart",
  "First": "Paul",
  "Last": "Stuart",
  "Email": "paul.stuart@pubmatic.com",
  "APIKey": "cc2c6a296f9f32017c30dd1b522d5532",
  "Level": 2
 }
xhr.status
200
            */
    /*
    attached: function() {
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
    }
    */

/*
var mLogin = Vue.component('okta-login', {
    template: '#tmpl-okta-login',
    mixins: [loginVue],
})
*/

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
/*
*/

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
    /*
        component: Vue.component('okta-login')
    */
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
    '/vlan/edit/:VLID': {
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
    '/part/add': {
        component:  Vue.component('part-edit')
    },
    '/part/edit/:PID': {
        component:  Vue.component('part-edit')
    },
    '/part/list': {
        component:  Vue.component('part-list')
    },
    '/part/inventory/:STI/:KID': {
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
    '/rma/edit/:RMAID': {
        component:  Vue.component('rma-edit')
    },
    '/rma/list': {
        component:  Vue.component('rma-list')
    },
    '/user/edit/:UID': {
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

