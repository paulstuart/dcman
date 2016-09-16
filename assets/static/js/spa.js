'use strict';


var userInfo = {};
var urlDebug = false;

var toggleDebug = function() {
    urlDebug = ! urlDebug;
    console.log('server-side debugging: ' + urlDebug);
}

var apikey = function() {
    if (userInfo && userInfo.APIKey && userInfo.APIKey.length > 0) {
        return userInfo.APIKey
    }
    return ""
}

var admin = function() {
    if (userInfo && userInfo.Level) {
        return userInfo.Level
    }
    return 0 
}

var get = function(url) {
  return new Promise(function(resolve, reject) {
    if (urlDebug) {
        url += (url.indexOf('?') > 0) ? '&debug=true' : '?debug=true'
    }
    // Do the usual XHR stuff
    var req = new XMLHttpRequest();
    req.open('GET', url);
    var key = apikey();
    if (key.length > 0) {
	    req.setRequestHeader("X-API-KEY", key)
    }

    req.onload = function() {
      // This is called even on 404 etc, so check the status
      //console.log('get status:', req.status, 'txt:', req.statusText)
      if (req.status == 200) {
        // Resolve the promise with the response text
        var obj = JSON.parse(req.responseText)
        resolve(obj)
      }
      else {
        // Otherwise reject with the status text
        // which will hopefully be a meaningful error
        console.log('rejecting!!! ack:',req.status, 'txt:', req.statusText)
        reject(Error(req.statusText));
      }
    };

    // Handle network errors
    req.onerror = function() {
      console.log('get network error');
      reject(Error("Network Error"));
    };

    // Make the request
    req.send();
  });
}

var posty = function(url, data, method) {
    return new Promise(function(resolve, reject) {
    // Do the usual XHR stuff
    if (typeof method == "undefined") method = 'POST';
    var req = new XMLHttpRequest();
    if (urlDebug) {
        url += (url.indexOf('?') > 0) ? '&debug=true' : '?debug=true'
    }
    req.open(method, url);
    var key = apikey();
    if (key.length > 0) {
	    req.setRequestHeader("X-API-KEY", key)
    }
	req.setRequestHeader("Content-Type", "application/json")

    req.onload = function() {
        // This is called even on 404 etc
        // so check the status
        console.log('get status:', req.status, 'txt:', req.statusText)
        if (req.status >= 200 && req.status < 300) {
            if (req.responseText.length > 0) {
                resolve(JSON.parse(req.responseText))
            } else {
                resolve(null)
            }
        }
        else {
            // Otherwise reject with the status text
            // which will hopefully be a meaningful error
            console.log('rejecting!!! ack:',req.status, 'txt:', req.statusText)
            if (req.getResponseHeader("Content-Type") === "application/json") {
                reject(JSON.parse(req.responseText));
            } else {
                reject(Error(req.statusText));
            }
        }
    };

    // Handle network errors
    req.onerror = function() {
        console.log('posty network error');
        reject(Error("Network Error"));
    };

    // Make the request
    req.send(JSON.stringify(data));
  });
}

// convenience wrapper
var deleteIt = function(url, fn) {
    if (fn) {
        posty(url, null, 'DELETE').then(fn)
    } else {
        posty(url, null, 'DELETE')
    }
}

function toQueryString(obj) {
    var parts = [];
    for (var i in obj) {
        if (obj.hasOwnProperty(i)) {
            parts.push(encodeURIComponent(i) + "=" + encodeURIComponent(obj[i]));
        }
    }
    return parts.join("&");
}

var postForm = function(url, data, fn, method) {
    var xhr = new XMLHttpRequest();
    if (typeof method == "undefined") var method = 'POST';
    var form = toQueryString(data);
    xhr.open(method, url, true);

    xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded; charset=UTF-8');

    xhr.send(form);
    xhr.onreadystatechange = function() {
        if (typeof fn === 'function') {
            fn(xhr);
            return
        }
        if (xhr.readyState == 4 && xhr.status != 200) {
            alert("Oops:" + xhr.responseText);
        }
    };
}


function getParameterByName(name, url) {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}

var buttonEnable = function(btn, enable) {
    if (enable) {
        btn.disabled = false
        btn.classList.remove("disabled")
    } else {
        btn.disabled = true
        btn.classList.add("disabled")
    }
}


var Maker = function(self, names, fresh) {
    self.Columns = function() {
        return names
    }

    self.Load = function(data) {
        for (var i=0; i < names.length; i++) {
            var column = names[i];
            self[column] = data[column]
        }
    }

    fresh = fresh || function(name) { return '' };

    self.Init = function() {
        for (var i=0; i < names.length; i++) {
            var column = names[i];
            self[column] = fresh(column)
        }
    }

    self.Init()
}

var newTable = function(name, template, mixins) {
  return Vue.component(name, {
      template: template,
      props: {
          data: Array,
          columns: Array,
          filterKey: String
      },
      data: function () {
          var sortOrders = {}
          this.columns.forEach(function (key) {
              sortOrders[key] = 1
          })
          return {
              sortKey: '',
              sortOrders: {} 
          }
      },
      methods: {
            sortBy: function (key) {
                this.sortKey = key
                this.sortOrders[key] = this.sortOrders[key] * -1
            },
      },
      mixins: mixins,
    })
}

var makeTable = function(template, mixins) {
  return Vue.extend({
      template: template,
      data: function () {
          return {
              columns: [],
              rows: [],
              sortKey: '',
              sortOrders: {},
              title: '',
          }
      },
      methods: {
            sortBy: function (key) {
                this.sortKey = key
                this.sortOrders[key] = this.sortOrders[key] * -1
            },
      },
      mixins: mixins,
      watch: {
          'columns': function(val, oldVal) {
              var self = this;
              this.columns.forEach(function (key) {
                  self.sortOrders[key] = 1
              })
          }
      },
    })
}

var makeNewTable = function(name, template, mixins) {
  return Vue.component(name, {
      template: template,
      data: function () {
          return {
              columns: [],
              rows: [],
              sortKey: '',
              sortOrders: {},
              title: '',
          }
      },
      methods: {
            sortBy: function (key) {
                this.sortKey = key
                this.sortOrders[key] = this.sortOrders[key] * -1
            },
      },
      mixins: mixins,
      watch: {
          'columns': function(val, oldVal) {
              var self = this;
              this.columns.forEach(function (key) {
                  self.sortOrders[key] = 1
              })
          }
      },
    })
}

var childTable = function(name, template, mixins) {
    return Vue.component(name, {
        template: template,
        props: [ 
              'columns',
              'rows',
              'filterKey',
              ],
        data: function() {
            var sortOrders = {}
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
            return {
                sortKey: '',
                sortOrders: sortOrders,
            }
        },
        methods: {
            sortBy: function (key) {
                this.sortKey = key
                this.sortOrders[key] = this.sortOrders[key] * -1
            },
        },
        mixins: mixins,
    })
}


var menuMIX = {
  data: {
      myapp: {
        auth: {
            loggedIn: true,
            user: {
                name: "Waldo"
            }
        },
      },
    }
}

var Site = function() {
    Maker(this, [
        'STI',
        'Name',
        'Address',
        'City',
        'State',
        'Country',
        'Web',
        'Note',
    ])
}

var RMA = function() {
    Maker(this, [
        'RMD',
        'DID',
        'STI',
        'VID',
        'OldPID',
        'NewPID',
        'Description',
        'Hostname',
        'ServerSN',
        'PartSN',
        'PartNumber',
        'VendorRMA',
        'Jira',
        'ShipTrack',
        'RecvTrack',
        'DCTicket',
        'Receiving',
        'Note',
        'Shipped',
        'Received',
        'Closed',
        'Created',
    ])
}

var Part = function() {
    Maker(this, [
        'PID',
        'DID',
        'STI',
        'PTI',
        'VID',
        'Site',
        'Hostname',
        'Description',
        'PartNumber',
        'Serial',
        'AssetTag',
        'Mfgr',
        'Price',
        'Cents',
        'Bad',
        'Used',
    ])
}

var Tag = function() {
    Maker(this, [
        'TID',
        'Name',
    ])
}

var VM = function() {
    Maker(this, [
        'VMI',
        'DID',
        'RID',
        'STI',
        'Rack',
        'Site',
        'Server',
        'Hostname',
        'Private',
        'Public',
        'Profile',
        'Note',
        'Version',
        'VIP',
    ])
}

var VLAN = function() {
    Maker(this, [
       'VLI',
       'STI',
       'Site',
       'Name',
       'Profile',
       'Gateway',
       'Route',
       'Netmask',
    ])
}

var User = function() {
    Maker(this, [
        'USR',
        'Login',
        'First',
        'Last',
        'Email',
        'Level',
    ])
}

var Rack = function() {
    Maker(this, [
        'RID',
        'STI',
        'Site',
        'RUs',
        'Label',
        'Note',
        'VendorID',
    ])
}

var Vendor = function() {
    Maker(this, [
        'VID',
        'Name',
        'WWW',
        'Phone',
        'Address',
        'City',
        'State',
        'Country',
        'Postal',
        'Note',
    ])
}

var Mfgr = function() {
    Maker(this, [
        'MID',
        'Name',
        'Note',
        'URL',
    ])
}

var IPType = function() {
    Maker(this, [
        'IPT',
        'Name',
        'Multi',
    ])
}

var DeviceType = function() {
    Maker(this, [
        'Name',
        'DTI',
    ])
}

var Device = function() {
    Maker(this, [
        'Alias',
        'AssetTag',
        'Assigned',
        'Site',
        'DID',
        'DTI',
        'Height',
        'Hostname',
        'MID',
        'Mfgr',
        'Model',
        'Note',
        'PartNo',
        'Profile',
        'Rack',
        'RID',
        'RU',
        'SerialNo',
        'STI',
        'TID',
        'Type',
        'Version',
    ])
}

var pingURL         = "http://10.100.182.16:8080/dcman/api/pings";
var macURL          = 'http://10.100.182.16:8080/dcman/data/server/discover/';

var adjustURL       = '/dcman/api/device/adjust/';
var assumeURL       = '/dcman/api/user/assume/'; 
var deviceAuditURL  = '/dcman/api/device/audit/';
var deviceListURL   = '/dcman/api/device/ips/';
var deviceNetworkURL = '/dcman/api/device/network/';
var deviceTypesURL  = '/dcman/api/device/type/';
var deviceURL       = '/dcman/api/device/view/';
var ifaceURL        = '/dcman/api/interface/';
var ifaceViewURL    = '/dcman/api/interface/view/';
var inURL           = "/dcman/api/inventory/";
var ipReserveURL    = '/dcman/api/network/ip/range';
var ipURL           = '/dcman/api/network/ip/';
var ipViewURL       = '/dcman/api/network/ip/view/';
var iptypesURL      = '/dcman/api/network/ip/type/';
var loginURL        = '/dcman/api/login';
var logoutURL       = '/dcman/api/logout';
var mfgrURL         = '/dcman/api/mfgr/';
var networkURL      = "/dcman/api/network/ip/used/";
var partTypesURL    = "/dcman/api/part/type/";
var partURL         = "/dcman/api/part/view/";
var rackURL         = "/dcman/api/rack/view/";
var reservedURL     = '/dcman/api/network/ip/reserved';
var rmaURL          = "/dcman/api/rma/";
var rmaviewURL      = "/dcman/api/rma/view/";
var searchURL       = "/dcman/api/search/";
var sessionsURL     = "/dcman/api/session/" ; 
var sitesURL        = "/dcman/api/site/" ; 
var summaryURL      = '/dcman/api/summary/';
var tagURL          = "/dcman/api/tag/";
var userURL         = "/dcman/api/user/" ; 
var vendorURL       = "/dcman/api/vendor/" ; 
var vlanURL         = "/dcman/api/vlan/view/" ; 
var vlanViewURL     = "/dcman/api/vlan/view/";
var vmAuditURL      = '/dcman/api/vm/audit/'
var vmIPsURL        = "/dcman/api/vm/ips/";
var vmURL           = "/dcman/api/vm/";
var vmViewURL       = "/dcman/api/vm/view/";

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
            if (id > 0) {
                url += query + id
            } else {
                url += query
            }
        } else if (id > 0) {
            url += id
        }
        return get(url).then(function(result) {
            //console.log('fetched:', what)
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
        if (all) {
            result.unshift({STI:0, Name:'All Sites'})
        }
        return result;
    })
    .catch(function(x) {
      console.log('Could not load sitelist: ', x);
    });
}

var getVendor = getIt(vendorURL, 'vendors');
var getPart = getIt(partURL, 'parts');
var getPartTypes = getIt(partTypesURL, 'part typess');
var getInventory = getIt(inURL, 'inventory');
var getDeviceTypes = getIt(deviceTypesURL, 'device typess');
var getIPTypes = getIt(iptypesURL, 'ip types');
var getDeviceLIST = getIt(deviceListURL, 'device list');
var getDeviceAudit = getIt(deviceAuditURL, 'device audit');
var getTagList = getIt(tagURL, 'tags');
var getDevice = getIt(deviceURL, 'device');
var getMfgr = getIt(mfgrURL, 'mfgr');
var getVM = getIt(vmViewURL, 'vm');
var getVMAudit = getIt(vmAuditURL, 'vm audit');
var getRack = getIt(rackURL, 'racks')
var getRMA = getIt(rmaviewURL, 'rma')
var getVLAN = getIt(vlanURL, 'vlan')
var getUser = getIt(userURL, 'user')
var getSessions = getIt(sessionsURL, 'sessions')


var foundLink = function(what) {
    switch (what.toLowerCase()) {
        case 'vm': return '/vm/edit/'
        case 'ip': return '/ip/edit/'
        case 'rack': return '/rack/edit/'
        case 'device': return '/device/edit/'
        case 'server': return '/device/edit/'
    }
    throw "Unknown link type: " + what;
}

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
            var port = ports[ifd]
            if (port.Mgmt > 0) {
                port.Port = 'IPMI'
            } else {
                port.Port = 'Eth' + port.Port
            }
            good.push(port)
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

var deviceVMs = function(device) {
    var url = vmURL + '?DID=' + device.DID; 
    return get(url).then(function(vms) {
        device.vms = vms
        return device
    })
}


// device info with associated interface / IPs
var completeDevice = function(DID) {
   return getDevice(DID).then(getInterfaces).then(deviceRacks).then(deviceVMs); 
}


var siteMIX = {
    route: { 
          data: function (transition) {
            return Promise.all([
                getSiteLIST(0), 
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
        posty(url + id, data, 'PATCH').then(fn)
    } else {
        posty(url, data).then(fn)
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
            pagerows: 10,
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

var authVue = {
    computed: {
        canEdit: function() {
            return (admin() > 0)
        },
        isAdmin: function() {
            return (admin() > 1)
        },
    },
}

// common stuff for edits 
var editVue = {
    mixins: [authVue],
    methods: {
        saveSelf: function() {
            var data = this.myself()
            var id = this.myID()
            if (id > 0) {
                posty(this.dataURL + id, data, 'PATCH').then(this.showList)
            } else {
                posty(this.dataURL, data).then(this.showList)
            }
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            posty(this.dataURL + this.myID(), null, 'DELETE').then(this.showList)
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


var fList = Vue.component('found-list', {
    template: '#tmpl-base-table',
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
            return foundLink(entry.Kind) + entry['ID']
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
})


// remove leading/trailing spaces, non-ascii
// TODO: perhaps just 'printable' chars?
var cleanText = function(text) {
    text = text.replace(/^[^A-Za-z0-9:\.\-]*/g, '')
    text = text.replace(/[^A-Za-z0-9:\.\-]*$/g, '')
    text = text.replace(/[^A-Za-z 0-9:\.\-]*/g, '')
    return text
}

var searchFor = Vue.component('search-for', {
    template: '#tmpl-search-for',
    data: function () {
        return {
            columns: ['Kind', 'Name', 'Note'],
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
            var url = searchURL + what;
            get(url).then(function(data) {
                console.log('we are searching for:', what)
                if (data) {
                    console.log('search matched:', data.length)
                    if (data.length == 1) {
                        console.log('what:', what);
                        var link = foundLink(data[0].Kind) + data[0].ID
                        router.go(link)
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
})


Vue.component('my-nav', {
    template: '#tmpl-main-menu',
    mixins: [authVue],
    props: ['app', 'msg'],
    data: function() {
       return {
           searchText: '',
           debug: false,
       }
    },
    created: function() {
        this.userinfo()
    },
    computed: {
        'debugAction': function() {
            return this.debug ? "Disable" : "Enable"
        },
    },
    methods: {
        'doSearch': function(ev) {
            var text = cleanText(this.searchText);
            if (text.length > 0) {
                console.log('initiate search for:',text)
                if (this.$route.name == 'search') {
                    // already on search page
                    this.$dispatch('search-again', text)
                    return
                }
                router.go({name: 'search', params: { searchText: text }})
            }
        },
        'toggleDebug': function() {
            this.debug = ! this.debug
            urlDebug = this.debug
            console.log('server-side debugging: ' + urlDebug);
        },
        'userinfo': function() {
            // reload existing auth from cookie
            var cookies = document.cookie.split("; ");
            var key = "";
            for (var i=0; i < cookies.length; i++) {
                var tuple = cookies[i].split('=')
                if (tuple[0] == 'X-API-KEY') {
                    key=tuple[1]
                    continue
                }
                if (tuple[0] != 'userinfo') continue;
                if (tuple[1].length == 0) break; // no cookie value so don't bother
                var user = JSON.parse(atob(tuple[1]));
                this.$dispatch('user-info', user, key)
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
            }

            get(url).then(function(data) {
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
            filename: 'iplist',
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
            return Promise.all([
                getSiteLIST(true), 
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
            return data.filter(function(obj) {
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

var reservedIPs = Vue.component('reserved-ips', {
    template: '#tmpl-reserved-ips',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            filename: 'iplist',
            rows: [],
            columns: [
                "Site",
                "VLAN",
                "IP",
                "Note",
                "User",
            ],
            sortKey: '',
            sortOrders: [],
            STI: 0,
            searchQuery: '',
            sites: [],
        }
    },
    route: { 
        data: function (transition) {
        var self = this;
        return Promise.all([
            getSiteLIST(true), 
            get(reservedURL),
        ]).then(function (data) {
            return {
                sites: data[0],
                rows: data[1],
            }
        })
        }
    },
    methods: {
        linkable: function(key) {
            return (key == 'IP')
        },
        linkpath: function(entry, key) {
            return '/ip/edit/' + entry['IID']
        },
    },
    filters: {
        siteFilter: function(data) {
            if (! this.STI > 0) {
                return data
            }

            var self = this;
            return data.filter(function(obj) {
                if (self.STI == obj.STI) {
                    return obj
                }
            });
        },
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
                posty(url + id, data, this.showList, 'PATCH')
            } else {
                posty(url + id, data, this.showList)
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
            columns: ['Login', 'First', 'Last', 'Level'],
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

            get(this.url).then(function(data) {
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
            current: userInfo,

            // TODO: pull levels data from server
            levels: [
                {Level:0, Label: 'User'},
                {Level:1, Label: 'Editor'},
                {Level:2, Label: 'Admin'},
            ],
        }
    },
    route: {
        data: function (transition) {
            if (transition.to.params.USR > 0) {
                return {
                    User: getUser(transition.to.params.USR)
                }
            }
            var user = new(User);
            user.USR = 0
            return {
                User: user
            }
        }
    },
    methods: {
        myID: function() {
            return this.User.USR
        },
        myself: function() {
            return this.User
        },
        canAssume: function() {
            return (this.User.USR > 0 && userInfo.Level > 1)
        },
        assumeUser: function() {
            var self = this;
            var url = assumeURL + this.User.USR;
            posty(url, null).then(function(user) {
                console.log("I AM NOW:", user);
                self.$dispatch('user-auth', user)
                router.go('/')
            })
        },
    },
})



var ipMIX = {
    mixins: [authVue],
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
            var ip  = row.IP
            var ipt = row.IPT
            var ifd = row.IFD
            var data = {IFD:ifd, IID: iid, IPT: ipt, IP: ip}
            console.log('update IP:', ip, ' IID:', iid)
            posty(ipURL + iid, data, null, 'PATCH')
            return false
        },
        deleteIP(i) {
            var self = this;
            var iid = this.rows[i].IID
            console.log("IP id:", iid)
            posty(ipURL + iid, null, 'DELETE').then(function() {
                self.rows.splice(i, 1)
            })
        },
        addIP: function() {
            var self = this;
            var data = {IFD: this.newIFD, IPT: this.newIPT, IP: this.newIP}
            console.log("we will add IP info:", data)
            posty(ipURL, data).then(function(ip) {
                self.rows.push(ip)
                self.newIP = ''
                self.newIPT = 0
                self.newIFD = 0
            })
            return false
        }
    }
}

var netgrid = childTable("network-grid", "#tmpl-network-grid", [ipMIX, authVue])

var interfaceMIX = {
    mixins: [authVue],
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

            var port = row.Port.replace(/[^\d]*/g, '');
            port = (port.length) ? parseInt(port) : 0

            var data = {
                IFD: ifd,
                Port: port,
                Mgmt: row.Mgmt,
                MAC: row.MAC,
                SwitchPort: row.SwitchPort,
                CableTag: row.CableTag,
            }

            posty(ifaceURL + ifd, data, null, 'PATCH')
        },
        deleteInterface(i) {
            var self = this;
            var ifd = this.rows[i].IFD
            console.log("Iface id:", ifd)
            posty(ifaceURL + ifd, null, 'DELETE').then(function() {
                self.rows.splice(i, 1)
            }).catch(function(ack) {
                console.log('=============> ACK!:', ack)
            })
        },
        addInterface: function(ev) {
            var self = this
            var port = this.newPort.replace(/[^\d]*/g, '');
            port = (port.length) ? parseInt(port) : 0
            var data = {
                DID: this.DID,
                Port: port,
                Mgmt: this.newMgmt,
                MAC: this.newMAC,
                SwitchPort: this.newSwitchPort,
                CableTag: this.newCableTag,
            }
            console.log("we will add interface info:", data)
            posty(ifaceURL, data).then(function(iface) {
                if (! iface.Mgmt) {
                    iface.Port = 'Eth' + iface.Port
                }
                self.rows.push(iface)

                self.newPort = ''
                self.newMgmt = ''
                self.newMAC = ''
                self.newSwitchPort = ''
                self.newCableTag = ''
            })
            return false
        }
    }
}

var ifacegrid = childTable("interface-grid", "#tmpl-interface-grid", [interfaceMIX])

//
// Device Edit
//
var deviceEditVue = {
    mixins: [editVue],
    data: function() {
        return {
            sites: [],
            device_types: [],
            mfgrs: [],
            tags: [],
            ipTypes: [],
            newIP: '',
            newIPD: 0,
            newIFD: 0,
            netColumns: ['IP', 'Type', 'Port'],
            ifaceColumns: ['Port', 'Mgmt', 'MAC', 'CableTag', 'SwitchPort'],
            Description: '',
            Device: new(Device),
            vmColumns: ['Hostname', 'Note'], 
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
                getMfgr(),
                completeDevice(this.$route.params.DID), 
           ]).then(function (data) {
              return {
                sites: data[0],
                device_types: data[1],
                tags: data[2],
                ipTypes:  data[3],
                mfgrs:  data[4],
                Device: data[5],
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

            if (device.DID == 0) {
                //device.Version = 0;
                console.log('save new device');
                posty(deviceURL, device).then(this.showList)
                return
            }
            console.log('update device id: ' + this.Device.DID);
            var url = deviceURL + this.Device.DID;
            posty(url, device, 'PATCH').then(this.showList)
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            var url = deviceURL + this.Device.DID;
            posty(url, device, 'DELETE').then(this.showList)
        },
        showList: function(ev) {
            router.go('/device/list')
        },
        portLabel: function(ipinfo) {
            if (this.Device.Type === 'server') {
                return ipinfo.Mgmt ? 'IPMI' : 'Eth' + ipinfo.Port
            } 
            return (ipinfo.Mgmt ? 'Mgmt' : 'Port') + ipinfo.Port
        },
        getMacAddr: function(ev) {
            var url = macURL + this.Server.IPIpmi;
            var self = this
            get(url).then(function(data) {
                self.Device.MacPort0 = data.MacEth0
                console.log("MAC DATA:", data)
             })
        },
        vmLinkable: function(key) {
            return (key == 'Hostname')
        },
        vmLinkpath: function(entry, key) {
            if (key == 'Hostname') return '/vm/edit/' + entry['VMI']
        },
        audit: function() {
            router.go('/device/audit/' + this.Device.DID)
        }
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
            device.DID     = 0;
            device.MID     = 0;
            device.TID     = 1;
            device.Height  = 1;
            device.Rack    = 0;
            device.Version = 0;
            device.STI     = parseInt(this.$route.params.STI);
            device.RID     = parseInt(this.$route.params.RID);
            device.RU      = parseInt(this.$route.params.RU);
            return {
                Device: deviceRacks(device)
            }
          },
    },
}

var deviceAdd = Vue.component('device-add', {
    template: '#tmpl-device-edit',
    mixins: [deviceEditVue, deviceAddMIX],
})


var deltas = function(ignore, data) {
    var rows = [];
    for (var i=0; i<data.length - 1; i++) {
        var before = data[i+1];
        var after  = data[i];
        Object.keys(before).forEach(function(key,index) {
            // key: the name of the object key
            // index: the ordinal position of the key within the object 
            if (! ignore.includes(key)) {
                var pre = before[key];
                var post = after[key];
                if (pre != post) {
                    var saved = {Column:key, Before: pre, After: post};
                    for (let keep of ignore) {
                        saved[keep] = after[keep]
                    }
                    rows.push(saved) 
                }
            }
        });
    }
    return rows
}

// audit data comes back with column and row data separate
// this will create our standard row with named fields
var deviceAudit = Vue.component('device-audit', {
    template: '#tmpl-device-audit',
    mixins: [pagedCommon],
    data: function() {
        return {
            filename: "audit",
            rows: [],
            columns: [ 
                "TS",
                "Version",
                "Login",
                "Column",
                "Before",
                "After",
            ],
        }
    },
    route: { 
        data: function (transition) {
            var self = this;
            var ignore = ["TS", "Version", "Login", "USR", "RID", "TID", "KID", "DTI"];
            return {
                rows: getDeviceAudit(transition.to.params.DID).then(function(fix) {
                    return deltas(ignore, fix)
                })
            }
        }
    },
    methods: {
        linkable: function(key) {
            //return (key == 'Name')
        },
        linkpath: function(entry, key) {
            //return '/ip/type/edit/' + entry['IPT']
        }
    },
    watch: {
        'STI': function(x) {
            this.loadData()
        }
    },
})


//
// VM IPs
//
var vmips = Vue.component('vm-ips', {
    template: '#tmpl-vm-ips',
    mixins: [authVue],
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
            var url = ipURL + '?VMI=' + this.VMI;
            get(url).then(function(data) {
                 self.rows = data
            })
            get(iptypesURL).then(function(data) {
                 self.types = data
                 console.log("IP TYPES:", data)
            })
        },
        updateIP(i) {
            var row = this.rows[i]
            var iid = row.IID
            var ip = row.IP
            var ipt = row.IPT
            var data = {VMI:this.VMI, IID: iid, IPT: ipt, IP: ip}
            console.log('update IP:', ip, ' IID:', iid)
            posty(ipURL + iid, data, null, 'PATCH')
            return false
        },
        deleteIP(i) {
            var self = this;
            var iid = this.rows[i].IID
            console.log("IP id:", iid)
            posty(ipURL + iid, null, 'DELETE').then(function() {
                self.rows.splice(i, 1)
            })
        },
        addIP: function() {
            var self = this;
            var data = {VMI: this.VMI, IPT: this.newIPT, IP: this.newIP}
            posty(ipURL, data, function(xhr) {
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
})


//
// VM Edit
//
var vmEdit = Vue.component('vm-edit', {
    template: '#tmpl-vm-edit',
    mixins: [authVue, siteMIX],
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
            posty(this.url + this.VM.VMI, this.VM, 'PATCH').then(this.showList)
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            posty(this.url + this.VM.VMI, null, 'DELETE').then(this.showList)
        },
        showList: function(ev) {
            router.go('/vm/list')
        },
    },
})


// audit data comes back with column and row data separate
// this will create our standard row with named fields
var vmAudit = Vue.component('vm-audit', {
    template: '#tmpl-audit',
    mixins: [pagedCommon],
    data: function() {
        return {
            filename: "audit",
            rows: [],
            columns: [ 
                "TS",
                "Version",
                "Login",
                "Column",
                "Before",
                "After",
            ],
        }
    },
    route: { 
        data: function (transition) {
            //var self = this;
            return {
                rows: getVMAudit(transition.to.params.VMI, "?vmi=").then(function(fix) {
                    var ignore = ["TS", "Version", "Login", "USR", "RID", "TID", "KID", "DTI"];
                    return deltas(ignore, fix)
                })
            }
        }
    },
    methods: {
        linkable: function(key) {
            //return (key == 'Name')
        },
        linkpath: function(entry, key) {
            //return '/ip/type/edit/' + entry['IPT']
        }
    },
})


// Base APP component, this is the root of the app
var App = Vue.extend({
    data: function(){
        return {
            myapp: {
                auth: {
                    // TODO: unify this with 'userInfo'
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
        'user-info': function (user, key) {
            this.myapp.auth.user.name = user.username;
            this.myapp.auth.user.admin = user.admin;
            this.myapp.auth.loggedIn = true;
            userInfo.Login = user.username;
            userInfo.Level = user.admin;
            userInfo.APIKey = key;
        },
        'user-auth': function (user) {
            console.log('*** user auth event:', user)
            this.myapp.auth.loggedIn = true;
            this.myapp.auth.user.name = user.Login;
            this.myapp.auth.user.admin = user.Level;
            userInfo = user;
        },
        'logged-out': function () {
            console.log('*** logged out event')
            this.myapp.auth.loggedIn = false
            this.myapp.auth.user.name = null
            this.myapp.auth.user.admin = 0
            userInfo = {};
            get(logoutURL)
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
        return {
            STI: 1,
            RID: 0,
            DTI: 0,
            sites: [],
            racks: [],
            searchQuery: '',
            rows: [],
            filename: "servers",
            types: [],
            columns: [
                 "Site",
                 "Rack",
                 "RU",
                 "Hostname",
                 "IPs",
                 "Mgmt",
                 "Type",
                 "Tag",
                 "Profile",
                 "Make",
                 "Model",
                 "SerialNo",
                 "AssetTag",
                 "Assigned",
                 "Note",
            ],
        }
    },
    filters: {
        rackFilter: function(data) {
            if (this.STI == 0) return data;
            if (this.RID == 0) return data;

            var self = this;
            return data.filter(function(obj) {
                if (obj.RID == self.RID) return obj
            });
        },
        typeFilter: function(data) {
            if (this.DTI == 0) return data;

            var self = this;
            return data.filter(function(obj) {
                if (obj.DTI == self.DTI) return obj
            });
        }
    },
    route: { 
          data: function (transition) {
            var self = this;
            console.log('device list promises starting for STI:', self.STI)
            return Promise.all([
                getSiteLIST(true), 
                self.STI > 0 ? getDeviceLIST(self.STI, '?sti=') : getDeviceLIST(), 
                self.STI > 0 ? getRack(self.STI, '?sti=') : getRack(), 
                getDeviceTypes(),
           ]).then(function (data) {
                console.log('device list promises returning')
                var racks =  data[2]
                racks.unshift({RID:0, Label:'All'})
             return {
                sites: data[0],
                rows: data[1],
                racks: data[2],
                types: data[3],
              }
            })
          }
    },
    methods: {
        reload: function() {
            this.RID = 0;
            var self = this;
            if (self.STI > 0) {
                getDeviceLIST(self.STI, '?sti=').then(function(devices) {
                    self.rows = devices
                })
                getRack(self.STI, '?sti=').then(function(racks) {
                    self.racks = racks
                })
            } else {
                getDeviceLIST().then(function(devices) {
                    self.rows = devices
                })
            }
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
// DEVICE TYPES
//

var deviceTypes = Vue.component('device-types', {
    template: '#tmpl-device-types',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
      console.log('device types returning')
      return {
          searchQuery: '',
          rows: [],
          columns: [
               "Name",
            ],
        }
    },
    route: { 
          data: function (transition) {
              return {
                  rows: getDeviceTypes()
              }
          }
    },
    methods: {
        addType: function() {
            router.go('/device/type/edit/0')
        },
        linkable: function(column) {
            return column === 'Name'
        },
        linkpath: function(entry, key) {
            if (key == 'Name') return '/device/type/edit/' + entry['DTI']
        }
    },
})

var deviceTypeEdit = Vue.component('device-type-edit', {
    template: '#tmpl-device-type-edit',
    data: function() {
        return {
            DeviceType: new(DeviceType)
        }
    },
    route: { 
          data: function (transition) {
              if (transition.to.params.DTI > 0) {
                  return {
                      DeviceType: getDeviceTypes(transition.to.params.DTI)
                  }
              }
              var dtype = new(DeviceType);
              dtype.DTI = 0;
              return {
                  DeviceType: dtype
              }
          }
    },
    methods: {
        saveSelf: function() {
            var self = this;
            var url = deviceTypesURL;
            if (this.DeviceType.DTI > 0) {
                url += this.DeviceType.DTI
                posty(url, this.DeviceType, 'PATCH').then(function() {
                    self.showList()
                });
            } else {
                posty(url, this.DeviceType).then(function() {
                    self.showList()
                })
            }
        },
        showList: function() {
            router.go('/device/types')
        },
        deleteSelf: function() {
            var url = deviceTypesURL + this.DeviceType.DTI
            posty(url, null, 'DELETE').then(this.showList);
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
            filename: "vlans",
            dataURL: vlanViewURL,
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
            get(url).then(function(data) {
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
            return data.filter(function(obj) {
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
            dataURL: vlanViewURL,
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
var ipReserve = Vue.component('ip-reserve', {
    template: '#tmpl-ip-reserve',
    mixins: [ authVue, siteMIX],
    data: function() {
        return {
            conflicted: 0,
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
            router.go('/ip/reserved')
        },
        reserveIPs: function() {
            var self = this;
            var url = ipReserveURL;
            var data = {
                From: this.From,
                To: this.To,
                VLI: this.VLI,
                Note: this.Note,
            }
            posty(url, data).then(this.showList).catch(function(fail) {
                var count=fail['Count'];
                if (count > 0) {
                    self.conflicted = count;
                }
                console.log("fail:", fail)
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
            get(url).then(function(data) {
                console.log('loaded vlan cnt:', data.length)
                self.vlans = data
            })
        },
        VLI: function() {
            console.log('VLI:', this.VLI, 'cnt:', this.vlans.length)
            for (var i=0; i < this.vlans.length; i++) {
                var vlan = this.vlans[i]
                if (vlan.VLI == this.VLI) {
                    var mask   = ip32(vlan.Netmask)
                    var net    = ip32(vlan.Gateway)
                    var min_ip = (net & mask) + 1
                    var max_ip = (net | ~mask) - 1

                    this.minIP32 = min_ip
                    this.maxIP32 = max_ip
                    this.Network = ipv4(min_ip) + ' - ' + ipv4(max_ip)
                    break
                }
            }
        }
    }
})


//
// IP EDIT
//
var ipEdit = Vue.component('ip-edit', {
    template: '#tmpl-ip-edit',
    mixins: [authVue],
    data: function() {
        return {
            IP: {},
            iptypes: [],
        }
    },
    computed: {
        'inuse': function() {
            return ((this.IP.VMI > 0) || (this.IP.IFD > 0))
        }
    },
    route: { 
        data: function (transition) {
            var url = ipViewURL + transition.to.params.IID;
            return {
                IP: get(url),
                iptypes: get(iptypesURL)
            }
        }
    },
    methods: {
        showList: function() {
            router.go('/ip/list')
        },
        saveSelf: function(event) {
            console.log('send update event: ' + event);
            var url = ipURL + this.IP.IID
            posty(url, this.IP, 'PATCH').then(this.showList)
        },
        deleteSelf: function() {
            var url = ipURL + this.IP.IID
            posty(url, this.IP, 'DELETE').then(this.showList)
        },
    },
})


//
// VM LIST
//

var vmList = Vue.component('vm-list', {
    template: '#tmpl-vm-list',
    mixins: [pagedCommon, siteMIX, commonListMIX],
    data: function() {
        return {
            filename: "vms",
            STI: 1,
            sites: [],
            site: 'blah',
            searchQuery: '',
            data: [],
            columns: [
                 "Site",
                 "Server",
                 "Hostname",
                 "IPs",
                 "Profile",
                 "Note",
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
             var self = this

             var url = vmIPsURL;
             if (self.STI > 0) {
                 url += "?sti=" + self.STI
             } 
             get(url).then(function(data) {
                 self.data = data
             })
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

var partInventory = Vue.component('part-inventory', {
    template: '#tmpl-part-inventory',
    mixins: [pagedCommon, commonListMIX, siteMIX],
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
            columns: ['Site', 'Description', 'PartNumber', 'Mfgr', 'Qty', 'Price'],
        }
    },
    route: {
        data: function (transition) {
            if (this.STI > 0) {
                return {
                    partData: getInventory(this.STI, '?sti='),
                    sites: getSiteLIST(true), 
                }
            }
            return {
                partData: getInventory(),
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
             if (this.STI > 0) {
                 getInventory(this.STI, '?sti=').then(function(data) {
                     self.partData = data
                 })
             } else {
                 getInventory().then(function(data) {
                     self.partData = data
                 })
             }
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
    computed: {
        'unusable': function() {
            return (((this.hostname.length == 0) || this.badHost) && this.other.length == 0);
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
            get(url).then(function(data) {
                self.available = data
            })
        },
        usePart: function(ev) {
            var pid = document.getElementById("PID").value
            var part = {
                PID: parseInt(pid),
                STI: this.STI,
                DID: this.DID,
                Unused: false,
            }
            posty(partURL + pid, part, "PATCH").then(this.showList)
        },
        findhost: function() {
            if (this.hostname.length === 0) {
                this.badHost = false
                return
            }
            var self = this;
            var url = deviceURL + "?hostname=" + this.hostname;
            get(url).then(function(hosts) {
                if (hosts && hosts.length == 1) {
                    self.DID = hosts[0].DID
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
            filename: "parts",
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
                var parts = data[1];
                if (parts) {
                    for (var i=0; i<parts.length; i++) {
                        parts[i].Price = parts[i].Price.toFixed(2);
                    }
                }
             return {
                sites: data[0],
                rows: parts,
                STI: STI,
              }
            })
        }
    },
    methods: {
      findhost: function(ev) {
          var self = this;
          console.log("find hostname:",this.hostname);
          get("api/server/hostname/" + this.hostname).then(function(data, status) {
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
            return data.filter(function(obj) {
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

var rmaList = Vue.component('rma-list', {
    template: '#tmpl-rma-list',
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            STI: 0,
            sites: [],
            rmas: [],
            searchQuery: '',
            rmaType: 1,
            columns: [
                "RMD",
                "Site",
                "Description",
                "Hostname",
                "PartSN",
                "VendorRMA",
                "Jira",
                "Created",
                "Shipped",
                "Received",
                "Closed",
            ],
            rmaType: 1,
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
             var url = rmaviewURL;
             if (self.STI > 0) {
                 url += "?STI=" + self.STI
             }
             get(url).then(function(data) {
                 if (! data) data = [];
                 self.rmas = data
             })
             getSiteLIST(true).then(function(sites) {
                 self.sites = sites
             })

        },
        linkable: function(key) {
            switch(key) {
                case 'Description': return true;
                case 'Hostname': return true;
            }
            return false;
        },
        linkpath: function(entry, key) {
            switch(key) {
                case 'Description': return '/rma/edit/' + entry['RMD']
                case 'Hostname': 
                    if (!('DID' in entry)) return '';
                    return '/device/edit/' + entry['DID']
            }
        },
    },
    filters: {
        rmaFilter: function(data) {
            if (this.rmaType == 1) {
                return data
            }
            var self = this;
            return data.filter(function(obj) {
                if (self.rmaType == 2 && ! obj.Closed) {
                    return obj
                }
                if (self.rmaType == 3 && obj.Closed) {
                    return obj
                }
            })
        },
    },
    watch: {
        'STI': function(val, oldVal){
                this.loadSelf()
            },
    },
})


//
// VENDOR LIST
//

var vendorList = Vue.component('vendor-list', {
    template: '#tmpl-vendor-list',
    mixins: [pagedCommon],
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
            get(vendorURL).then(function(data) {
                self.rows = data
            })
        },
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/vendor/edit/' + entry['VID']
        },
    },
})


var vendorEdit = Vue.component('vendor-edit', {
    template: '#tmpl-vendor-edit',
    mixins: [editVue],
    data: function() {
        var vendor = new(Vendor);
        vendor.VID = 0
        return {
            Vendor: vendor,
            dataURL: vendorURL,
            listURL: '/vendor/list',
        }
    },
    route: { 
        data: function (transition) {
            if (this.$route.params.VID > 0) {
                return {
                    Vendor: getVendor(this.$route.params.VID)
                }
            }
            var vendor = new(Vendor);
            vendor.VID = 0
            return {
                Vendor: vendor
            }
          },
    },
    methods: {
        myID: function() {
            return this.Vendor.VID
        },
        myself: function() {
            return this.Vendor
        },
        showList: function() {
            router.go('/vendor/list')
        },
    },
})


//
// MFGR LIST
//

var mfgrList = Vue.component('mfgr-list', {
    template: '#tmpl-mfgr-list',
    mixins: [authVue],
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
            get(mfgrURL).then(function(data) {
                self.rows = data
            })
        },
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/mfgr/edit/' + entry['MID']
        },
    },
})



var mfgrEdit = Vue.component('mfgr-edit', {
    template: '#tmpl-mfgr-edit',
    mixins: [authVue],
    data: function() {
        return {
            Mfgr: {},
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
            if (transition.to.params.MID > 0) {
                return {
                    Mfgr: getMfgr(transition.to.params.MID)
                }
            }
            var mfgr = new(Mfgr);
            mfgr.MID = 0
            return {
                Mfgr: mfgr,
            }
        }
    },
})


//
// PART EDIT
//
var partEdit = Vue.component('part-edit', {
    mixins: [authVue],
    template: '#tmpl-part-edit',
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
            if (this.Part.PTI == 0) {
                return (this.Part.STI == 0 || this.Part.PTI == 0 || this.Part.Description.length == 0)
            }
        },
        badPrice: function() {
            return false
        }
    },
    route: { 
        data: function (transition) {
            return {
                Part: getPart(transition.to.params.PID),
                sites: getSiteLIST(),
                types: getPartTypes(),
                vendors: getVendor().then(function(list) {
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
                posty(url, this.Part, 'PATCH').then(this.showList)
            } else {
                posty(url, this.Part).then(this.showList)
            }
        },
        doRMA: function(ev) {
            router.go('/rma/create/' + this.Part.PID)
        },
        findhost: function() {
            if (this.Part.Hostname.length === 0) {
                this.Part.DID = 0
                this.badHost = false
                return
            }
            var self = this;
            var url = deviceURL + "?hostname=" + this.Part.Hostname;
            get(url).then(function(hosts) {
                if (hosts && hosts.length == 1) {
                    self.Part.DID = hosts[0].DID
                    self.badHost = false;
                } else {
                    self.badHost = true;
                }
            })
        },
    },
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
                posty(rmaviewURL + this.RMA.RMD, this.RMA, 'PATCH').then(this.showList)
            } else {
                posty(rmaviewURL, this.RMA).then(this.showList)
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
            get(url).then(function(hosts) {
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

//
// RMAs
//

var rmaEdit = Vue.component('rma-edit', {
    template: '#tmpl-rma-edit',
    mixins: [authVue, rmaCommon],
    route: { 
        data: function (transition) {
            return {
                RMA: getRMA(this.$route.params.RMD),
            }
        }
    },
    methods: {
        deleteSelf: function(event) {
            var url = this.dataURL + this.myID();
            posty(url, null, 'DELETE').then(this.showList)
        },
    },
})


//
// RMA CREATE
//

var rmaCreate = Vue.component('rma-create', {
    template: '#tmpl-rma-edit',
    mixins: [rmaCommon, authVue],
    route: { 
        data: function (transition) {
            var part = getPart(this.$route.params.PID);

            var now = new Date();
            var dd = now.getDate();
            var mm = now.getMonth()+1; //January is 0!
            var yyyy = now.getFullYear();
            var today = yyyy + '/' + mm + '/' + dd;

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
                        rma.Created = today
                        return(rma)
                }),
            }
        },
    },
})


//
// PART LOAD
//
//
var partLoad = Vue.component('part-load', {
    template: '#tmpl-part-load',
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
            // normalize column names
            var partCol = function(col) {
                switch (col) {
                    case "Item":            return "PartType";
                    case "Part Number":     return "PartNumber";
                    case "Manufacturer":    return "Mfgr";
                    case "Cost":            return "Price";

                    case "SKU":             
                    case "Description":     
                    case "Qty":             
                    case "Price":           return col;
                }
            }
            var parts = this.Parts.split("\n");
            var cols = {};
            for (var i=0; i < parts.length; i++) {
                var line = parts[i].split("\t");

                // parse column headers
                if (i === 0) {
                    for (var k=0; k < line.length; k++) {
                        var col = partCol(line[k])
                        if (col) {
                            cols[k] = col
                        }
                    }
                    console.log("COLS:", cols)
                    continue
                }

                var part = new(Part);
                part.PID = null;
                part.KID = null;
                part.DID = null;
                part.VID = null;
                part.STI = this.STI;
                part.Bad = false;
                part.Unused = true;
                for (var j in cols) {
                    var col = cols[j];
                    part[col] = line[j];
                }
                var qty = parseInt(part["Qty"]);
                if (qty === 0) qty = 1;
                if (part.Price) {
                    part.Price = part.Price.replace(/[^0-9.]*/g,'')
                    part.Price = parseFloat(part.Price)
                    part.Cents = Math.round(part.Price * 100)
                } else {
                    part.Price = 0.0
                    part.Cents = 0
                }
                var url = partURL;
                for (var j=0; j < qty; j++) {
                    posty(url, part)
                }
            }
            this.showList()
        },
    },
})


//
// TAGS
//

var tagEdit = Vue.component('tag-edit', {
    template: '#tmpl-tag-edit',
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
            return {
              sites: getSiteLIST(), 
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
             get(this.url).then(function(data) {
                 self.tags = data
             })
        },
        deleteSelf: function(ev) {
            console.log("delete self...")
            if (! this.tag) {
                return
            }
            console.log('delete tag url: ' + this.url + this.tag.TID)
            posty(this.url + this.tag.TID, null, function(data) {}, 'DELETE')
            console.log("delete tid:",this.tag.TID)
            for (var i=0; i < this.tags.length; i++) {
                console.log("i:",i,"tid:",this.tags[i].TID)
                if (this.tags[i].TID == this.tag.TID) {
                    console.log("deleting tag:", i, "of", this.tags.length)
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
                var self = this
                var refresh = function() {
                    for (var i=0; i < self.tags.length; i++) {
                        if (self.tags[i].TID == self.tag.TID) {
                            self.tags[i].Name = self.tag.Name
                            break
                        }
                    }
                }
                posty(this.url + this.tag.TID, this.tag, refresh, 'PATCH')
            } else {
                posty(this.url, this.tag, saved)
            }
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
})


//
// RACK Edit
//
var rackEdit = Vue.component('rack-edit', {
    template: '#tmpl-rack-edit',
    mixins: [editVue, siteMIX],
    data: function() {
        var rack = new(Rack);
        rack.RID = 0;
        return {
            sites: [],
            id: 'RID',
            name: 'Rack',
            Rack: rack,
            dataURL: rackURL,
            listURL: '/rack/list',
        }
    },
    computed: {
        notReady: function() {
            if (this.Rack.STI < 1) return true
            if (this.Rack.Label < 1) return true
            if (parseInt(this.Rack.RUs) < 1) return true
            return false
        },
    },
    route: {
        data: function (transition) {
            if (transition.to.params.RID > 0) {
                return {
                    Rack: getRack(transition.to.params.RID)
                }
            }
            var rack = new(Rack);
            rack.RID = 0
            rack.STI = 0
            return {
                Rack: rack 
            }
        },
    },
    methods: {
          showList: function(ev) {
              router.go('/rack/list')
          },
        myself: function() {
            return this.Rack
        },
        myID: function() {
            return this.Rack.RID
        }
    },
})


//
// RACK LIST
//

var rackList = Vue.component('rack-list', {
    template: '#tmpl-rack-list',
    mixins: [pagedCommon, siteMIX, commonListMIX],
    data: function() {
        return {
            dataURL: rackURL,
            STI: 0,
            sites: [],
            searchQuery: '',
            rows: [],
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
             get(url).then(function(data) {
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
    var unracked = {}
    for (var k=0; k<racks.length; k++) {
        var rack = racks[k]
        lookup[rack.RID] = rack

        // pre-populate empty rack
        var these = [];
        var size = rack.RUs;
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
            if (unit.RU > 0) {
                byRID[unit.RID][rack.RUs - unit.RU] = unit
            } else {
                // PDUs and other rack items w/o an RU
                if (! unracked[unit.RID]) {
                    unracked[unit.RID] = []
                }
                unracked[unit.RID].push(unit)
            }
        }
    }

    var lumps = []
    for (var i=0; i<racks.length; i++) {
        var rack = racks[i];
        if (! rack || rack.RID == 0) continue
        var these = byRID[rack.RID]

        // for units greater than 1 RU, hide the slots consumed above
        // work our way up from the bottom
        for (var k=these.length - 1; k > 0; k--) {
            for (var j=these[k].Height; j > 1; j--) {
                var x = k-j+1
                if (!these[x]) { continue }
                these[x].Height = 0;
            }
            these[k]['pingMgmt'] = ''
            these[k]['pingIP'] = ''
        }
        lumps.push({rack: rack, units: these, other: unracked[rack.RID]})
    }
    return lumps
}


// for rack-layout
var rackView = Vue.component('rack-view', {
    template: '#tmpl-rack-view',
    props: ['layouts', 'RID', 'audit'],
    data: function() {
        return {
            fields: "newHeight newHostname newIP newMgmt Height Hostname DID RID Mgmt IPs".split(" "),
        }
    },
    methods: {
        rfilter: function(a, b, c) {
            if (this.RID == 0) {
                return a
            }
            if (this.RID == this.rack.RID) {
                return a
            }
        },
        move: function(lay, up) {
            console.log('move event:', lay)
        },
        copy: function(lay, ru) {
            var off = this.layouts.rack.RUs - ru;
            var device = this.layouts.units[off];
            for (var i=0; i < this.fields.length; i++) {
                var f = this.fields[i];
                device[f] = lay[f]
            }
        },
        one: function(ru) {
            var off = this.layouts.rack.RUs - ru;
            var device = this.layouts.units[off];
            device.newHeight = 1
            device.Height    = 1
        },
        rusize: function(ru, size) {
            var off = this.layouts.rack.RUs - ru;
            var device = this.layouts.units[off];
            device.newHeight = size
            device.Height    = size
        },
        zero: function(ru) {
            var off = this.layouts.rack.RUs - ru;
            var device = this.layouts.units[off];
            device.newHeight    = 0
            device.newHostname  = ''
            device.Height    = 0
            device.Hostname  = ''
            device.DID       = 0
            device.RID       = 0
            device.Mgmt      = ''
            device.IPs       = ''
        },
        moveUp: function(lay) {
            var ru = lay.RU + 1;
            var self = this;
            var url = adjustURL + lay.DID;
            var adjust = {DID: lay.DID, RID: lay.RID, RU: ru, Height: lay.Height};
            posty(url, adjust, 'PUT').then(function(moved) {
                if (moved.RU == ru) {
                    if (lay.Height > 1) self.zero(lay.RU + lay.Height)
                    self.copy(lay, ru)
                    self.zero(lay.RU)
                    self.rusize(lay.RU, 1)
                }
            })
        },
        moveDown: function(lay) {
            var ru = lay.RU - 1;
            var self = this;
            var url = adjustURL + lay.DID;
            var adjust = {DID: lay.DID, RID: lay.RID, RU: ru, Height: lay.Height};
            posty(url, adjust, 'PUT').then(function(moved) {
                if (moved.RU == ru) {
                    self.copy(lay, ru)
                    if (lay.Height > 1) {
                        self.rusize(ru + lay.Height, 1)
                    } else {
                        self.rusize(lay.RU, 1)
                    }
                    self.zero(lay.RU)
                }
            })
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
                var ip = {IID: ipinfo.IID, IP: lay.newIP}
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
                if (! data || data.length != 1) {
                    console.log('bad mgmt data:', data)
                    return
                }
                var ipinfo = data[0]
                console.log("IPINFO:", ipinfo);
                var ip = {IID: ipinfo.IID, IP: lay.newMgmt}
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
            var url = adjustURL + lay.DID;
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
                    return (unit.RU + unit.Height) < ru;
                }
            }
            return true
        }
    }
})


//
// RACK LAYOUT
//

var rackLayout = Vue.component('rack-layout', {
    template: '#tmpl-rack-layout',
    mixins: [authVue, commonListMIX],
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
            return {
                sites: getSiteLIST()
            }
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
             } else if (self.STI > 0) {
                 url += "?sti=" + self.STI
             }

             get(url).then(function(units) {
                 url = rackURL + "?STI=" + self.STI;

                 get(url).then(function(racks) {
                     if (racks) {
                         racks.unshift({RID:0, Label:''})
                         self.racks = racks
                         self.lumpy = makeLumps(racks, units)
                     }
                 })
             })
        },
        ping: function() {
            var url = pingURL;
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
                this.RID = 0;
                this.audit = false;
                this.loadSelf()
        },
        'RID': function(val, oldVal) {
                console.log('RID is now:', val)
        },
    },
})


var userLogin = Vue.component('user-login', {
    template: '#tmpl-user-login',
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
            var data = {Username: this.username, Password: this.password};
            var self = this;
            posty(loginURL, data).then(function(user) {
                self.$dispatch('user-auth', user)
                router.go('/')
            }).catch(function(msg) {
                self.errorMsg = msg.Error
            })
        },
    },
})


var userLogout = Vue.component('user-logout', {
    template: '#tmpl-user-logout',
    methods: {
        cancel: function() {
            router.go('/')
        },
        logout: function(ev) {
            this.$dispatch('logged-out')
            router.go('/')
        },
    }
})


// grid component with paging and sorting
var pagedGrid = Vue.component('paged-grid', {
    template: '#tmpl-paged-grid',
    props: {
        data: Array,
        columns: Array,
        linkable: Function,
        linkpath: Function,
        startRow: Number,
        rowsPerPage: Number,
        filename: String,
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
            var status = 
                ' Page ' +
                (this.startRow / this.rowsPerPage + 1) +
                ' / ' +
                (Math.ceil(this.data.length / this.rowsPerPage));

            if (this.data.length >  this.rowsPerPage) {
                status += " (" + this.data.length + " rows) ";
            }
            return status
        },
        canDownload: function() {
            /*
            console.log("DATA OK:", (this.data && (this.data.length > 0)));
            console.log("FILE OK:", (filename && (filename.length > 0)));
            */
            return (this.data && (this.data.length > 0) && this.filename && (this.filename.length > 0))
        },
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
        download() {
            // TODO: perhaps get fancier and use this?
            // https://github.com/eligrey/FileSaver.js#saving-text
            var filename = this.filename;
            if (filename.indexOf(".") < 0 ) {
                filename += ".xls";
            }

            // gather up our data to save (tab delimited)
            var text = this.columns.join("\t") + "\n";
            for (var i=0; i < this.data.length; i++) {
                var line = [];
                for (var j=0; j < this.columns.length; j++) {
                    var col = this.columns[j];
                    line.push(this.data[i][col])
                }
                text += line.join("\t") + "\n";
            }

            var element = document.createElement('a');
            var ctype = 'application/vnd.ms-excel';
            element.setAttribute('href', 'data:' + ctype + ';charset=utf-8,' + encodeURIComponent(text));
            element.setAttribute('download', filename);

            element.style.display = 'none';
            document.body.appendChild(element);

            element.click();

            document.body.removeChild(element);
        },
    }
});


var sessionList = Vue.component('session-list', {
    template: '#tmpl-session-list',
    mixins: [pagedCommon],
    data: function() {
        return {
            filename: "sessions",
            columns: ['TS', 'Login', 'Remote', 'Event'],
            rows: [],
        }
    },
    route: { 
        data: function (transition) {
            return {
              rows: getSessions(), 
            }
        }
    },
    methods: {
        linkable: function(key) {
            return (key == 'Login')
        },
        linkpath: function(entry, key) {
            return '/user/edit/' + entry['USR']
        }
    }
})


//
// SITE LIST
//
var siteList = Vue.component('site-list', {
    template: '#tmpl-site-list',
    mixins: [authVue],
    data: function() {
        return {
            sites: [],
            searchQuery: '',
            rows: [],
            columns: [
                'Name',
                'City',
                'Country',
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
            get(sitesURL).then(function(data) {
                self.rows = data
            })
        },
        linkable: function(key) {
            return (key == 'Name')
        },
        linkpath: function(entry, key) {
            return '/site/edit/' + entry['STI']
        },
    },
})

var siteEdit = Vue.component('site-edit', {
    template: '#tmpl-site-edit',
    mixins: [authVue],
    data: function() {
        return {
            Site: {},
        }
    },
    methods: {
        showList: function() {
            router.go('/site/list')
        },
        deleteSelf: function() {
            deleteIt(sitesURL + this.Site.STI, this.showList)
        },
        saveSelf: function()  {
            saveMe(sitesURL, this.Site, this.Site.STI, this.showList)
        } 
    },
    route: { 
        data: function (transition) {
            if (transition.to.params.STI > 0) {
                var url = sitesURL + transition.to.params.STI;
                return {
                    Site: get(url),
                }
            }
            var site = new(Site);
            site.STI = 0
            return {
                Site: site,
            }
        }
    },
})



var homePage = Vue.component('home-page', {
    template: '#tmpl-home-page',
    data: function() {
        return {
            title: "PubMatic Datacenters",
            rows: [],
            columns: [ "Site", "Servers", "VMs" ],
            testData: 'this is a test',
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            var self = this;
            get(summaryURL).then(function(data) {
                self.rows = data
            })
        },
        dl: function() {
            download('test.txt', this.testData)
        }
    },
})

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
    '/admin/sessions': {
        component: Vue.component('session-list')
    },
    '/admin/tags': {
        component: Vue.component('tag-edit')
    },
    '/ip/edit/:IID': {
        component:  Vue.component('ip-edit')
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
    '/ip/reserved': {
        component:  Vue.component('reserved-ips')
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
    '/device/audit/:DID': {
        component: Vue.component('device-audit')
    },
    '/device/edit/:DID': {
        component: Vue.component('device-edit')
    },
    '/device/list': {
        component:  Vue.component('device-list')
    },
    '/device/types': {
        component:  Vue.component('device-types')
    },
    '/device/type/edit/:DTI': {
        component:  Vue.component('device-type-edit')
    },
    '/vm/audit/:VMI': {
        component: Vue.component('vm-audit')
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
    '/site/edit/:STI': {
        component:  Vue.component('site-edit')
    },
    '/site/list': {
        component:  Vue.component('site-list')
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
    if ((! userInfo.APIKey || userInfo.APIKey.length == 0) && transition.to.path !== '/auth/login') {
        router.go('/auth/login')
    } else {
        transition.next()
    }
})


router.start(App, '#app')
