"use strict";

console.log("loading spa.js");

var fromCookie = function() {
    // reload existing auth from cookie
    for (var cookie of document.cookie.split("; ")) {
        const tuple = cookie.split("=")
        if (tuple[0] == "userinfo") {
            if (tuple[1].length > 0) {
		try {
			return JSON.parse(atob(tuple[1]));
		} catch(e) {
			console.log("cookie parse error:", e)
		}
            }
        }
    }
    return null
}

var killCookie = function() {
    var xhttp = new XMLHttpRequest();
    xhttp.open("GET", "api/logout", true);
    xhttp.send();
}

var sameUser = false;

const store = new Vuex.Store({
    state: {
        apiKey: "",
        level: 0,
        active: false,
        login: "",
        USR: null,
    },
    getters: {
        canEdit: state => {
            return (state.level > 0)
        },
        isAdmin: state => {
            return (state.level > 1)
        },
        userName: state => {
            return state.login
        },
        apiKey: state => {
            return state.apiKey
        },
        loggedIn: state => {
            return state.active
        },
    },
    mutations: {
        setUser(state, user) {
            state.level = user.Level
            state.apiKey = user.APIKey
            state.login = user.Login
            state.USR = user.USR
            state.active = true
         //   console.log("setUser:", state.login);
            if (user["COOKIE"] === true) {
        //        console.log("verifying cookie data is valid");
                get("api/check")
/*
                    .then(u => 
                        console.log("user is good:", u)
                    )
*/
                    .catch(x => {
                        console.log("user is bad:", x)
                    })
            }
        },
        logOut(state) {
            console.log("logging out:", state.login);
            state.level = 0
            state.apiKey = ""
            state.login = ""
            state.active = false
            state.USR = null
        },
    },
})

const authVue = {
    computed: {
        canEdit: function() {
            return this.$store.getters.canEdit
        },
    },
    created: function() {
        // needed if refreshing a page and session expired
        if (! this.$store.getters.loggedIn) {
            console.log("NOT LOGGED IN!");
            router.push("/auth/login")
        }
    },
}

var urlDebug = false;

var toggleDebug = function() {
    urlDebug = ! urlDebug;
    console.log("server-side debugging: " + urlDebug);
}

var get = function(url) {
  return new Promise(function(resolve, reject) {
    if (urlDebug) {
        url += ((url.indexOf("?") > 0) ? "&" : "?") + "debug=true"
    }
    // Do the usual XHR stuff
    var req = new XMLHttpRequest();
    req.open("GET", url);
    var key = store.getters.apiKey;
    if (key.length > 0) {
        req.setRequestHeader("X-API-KEY", key)
    } else {
        router.push("/auth/login")
        //reject(Error("no api key set"));
        return
        //console.log("no api key set")
    }

    req.onload = function() {
      // This is called even on 404 etc, so check the status
      if (req.status == 200) {
        // Resolve the promise with the response text
        var obj = (req.responseText.length > 0) ? JSON.parse(req.responseText) : null;
        resolve(obj)
      }
      else {
          if (req.status == 401) {
              store.commit("logOut")
              //  resolve(null)
              reject(Error(req.statusText));
                killCookie();
               router.go("/auth/login")
              return
          }
          // Otherwise reject with the status text
          // which will hopefully be a meaningful error
          console.log("rejecting!!! ack:",req.status, "txt:", req.statusText)
          reject(Error(req.statusText));
      }
    };

    // Handle network errors
    req.onerror = function() {
      console.log("get network error");
      reject(Error("Network Error"));
    };

    // Make the request
    req.send();
  });
}

var posty = function(url, data, method) {
    return new Promise(function(resolve, reject) {
    // Do the usual XHR stuff
    if (typeof method == "undefined") method = "POST";
    var req = new XMLHttpRequest();
    if (urlDebug) {
        url += (url.indexOf("?") > 0) ? "&debug=true" : "?debug=true"
    }
    req.open(method, url)
    const key = store.getters.apiKey;
    if (key.length > 0) {
        req.setRequestHeader("X-API-KEY", key)
    }
    req.setRequestHeader("Content-Type", "application/json")

    req.onload = function() {
        // This is called even on 404 etc
        // so check the status
        //console.log("get status:", req.status, "txt:", req.statusText)
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
            console.log("rejecting!!! ack:",req.status, "txt:", req.statusText)
            if (req.getResponseHeader("Content-Type") === "application/json") {
                reject(JSON.parse(req.responseText));
            } else {
                reject(Error(req.statusText));
            }
        }
    };

    // Handle network errors
    req.onerror = function() {
        console.log("posty network error");
        reject(Error("Network Error"));
    };

    // Make the request
    req.send(JSON.stringify(data));
  });
}

// convenience wrapper
var deleteIt = function(url, fn) {
    if (fn) {
        posty(url, null, "DELETE").then(fn)
    } else {
        posty(url, null, "DELETE")
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
    if (typeof method == "undefined") var method = "POST";
    var form = toQueryString(data);
    xhr.open(method, url, true);

    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8");

    xhr.send(form);
    xhr.onreadystatechange = function() {
        if (typeof fn === "function") {
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
    if (!results[2]) return "";
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


var childTable = function(name, template, mixins) {
    return Vue.component(name, {
        template: template,
        props: [
            "columns",
            "rows",
            "filterKey",
        ],
        data: function() {
            var sortOrders = {}
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
            return {
                sortKey: "",
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


var pingURL         = "api/ping";
var macURL          = "data/server/discover/";

var adjustURL       = "api/device/adjust/";
var assumeURL       = "api/user/assume/";
var deviceAuditURL  = "api/device/audit/";
var deviceListURL   = "api/device/ips/";
var deviceNetworkURL = "api/device/network/";
var deviceTypesURL  = "api/device/type/";
var deviceURL       = "api/device/";
var deviceMacURL    = "api/device/mac/";
var deviceViewURL   = "api/device/view/";
var ifaceURL        = "api/interface/";
var ifaceViewURL    = "api/interface/view/";
var inURL           = "api/inventory/";
var ipReserveURL    = "api/network/ip/range";
var ipURL           = "api/network/ip/";
var ipViewURL       = "api/network/ip/view/";
var iptypesURL      = "api/network/ip/type/";
var loginURL        = "api/login";
var logoutURL       = "api/logout";
var mfgrURL         = "api/mfgr/";
var networkURL      = "api/network/ip/used/";
var partTypesURL    = "api/part/type/";
var partURL         = "api/part/view/";
var rackURL         = "api/rack/";
var rackViewURL     = "api/rack/view/";
var reservedURL     = "api/network/ip/reserved";
var rmaURL          = "api/rma/";
var rmaviewURL      = "api/rma/view/";
var searchURL       = "api/search/";
var sessionsURL     = "api/session/" ;
var sitesURL        = "api/site/" ;
var summaryURL      = "api/summary/";
var tagURL          = "api/tag/";
var userURL         = "api/user/" ;
var vendorURL       = "api/vendor/" ;
var vlanURL         = "api/vlan/" ;
var vlanViewURL     = "api/vlan/view/";
var vmAuditURL      = "api/vm/audit/"
var vmIPsURL        = "api/vm/ips/";
var vmURL           = "api/vm/";
var vmViewURL       = "api/vm/view/";

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
            return result;
        })
        .catch(function(x) {
            console.log("fetch failed for:", what, "because:", x);
            throw(x)
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
            result.unshift({STI:0, Name:"All Sites"})
        }
        return result;
    })
    .catch(function(x) {
      console.log("Could not load sitelist: ", x);
    });
}

var getVendor = getIt(vendorURL, "vendors");
var getPart = getIt(partURL, "parts");
var getPartTypes = getIt(partTypesURL, "part types");
var getInventory = getIt(inURL, "inventory");
var getDeviceTypes = getIt(deviceTypesURL, "device types");
var getIPTypes = getIt(iptypesURL, "ip types");
var getDeviceLIST = getIt(deviceListURL, "device list");
var getDeviceAudit = getIt(deviceAuditURL, "device audit");
var getTagList = getIt(tagURL, "tags");
var getDevice = getIt(deviceViewURL, "device");
var getMfgr = getIt(mfgrURL, "mfgr");
var getVM = getIt(vmViewURL, "vm");
var getVMAudit = getIt(vmAuditURL, "vm audit");
var getRack = getIt(rackViewURL, "racks")
var getRMA = getIt(rmaviewURL, "rma")
var getVLAN = getIt(vlanURL, "vlan")
var getUser = getIt(userURL, "user")
var getSessions = getIt(sessionsURL, "sessions")


var foundLink = function(what) {
    switch (what.toLowerCase()) {
        case "vm": return "/vm/edit/"
        case "ip": return "/ip/edit/"
        case "rack": return "/rack/edit/"
        case "device": return "/device/edit/"
        case "server": return "/device/edit/"
    }
    throw "Unknown link type: " + what;
}

var getInterfaces = function(device) {
    var url = ifaceViewURL + "?DID=" + device.DID;
    return get(url).then(function(interfaces) {
        if (! interfaces) {
            device.ips = []
            device.interfaces = []
            return device
        }
        var ips = []
        var ports = {}
        interfaces.forEach(iface => {
            if (! (iface.IFD in ports)) {
                ports[iface.IFD] = iface
            }
            if (iface.IP) ips.push(iface)
        })
        var good = [];
        Object.keys(ports).map(key => {
            var port = ports[key]
            if (port.Mgmt > 0) {
                port.Port = "IPMI"
            } else {
                port.Port = "Eth" + port.Port
            }
            good.push(port)
        })
        device.ips = ips
        device.interfaces = good
        return device
   })
}

var deviceRacks = function(device) {
    var url = rackViewURL + "?STI=" + device.STI;
    return get(url).then(function(racks) {
        device.racks = racks
        return device
    })
}

var deviceVMlist = function(device) {
    var url = vmIPsURL + "?DID=" + device.DID;
    return get(url).then(function(v) {
        device.vms = v;
        return device
    })
}

var siteMIX = {
    created: function() {
        getSiteLIST(true).then(s => this.sites = s);
    }
}

var validIP = function(ip) {
    if (! ip || ip.length === 0) return false;
    var octs = ip.split(".")
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
    var octs = ip.split(".");
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
        posty(url + id, data, "PATCH").then(fn)
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
            searchQuery: "",
            startRow: 0,
            pagerows: 10,
            sizes: [10, 25, 50, 100, "all"],
        }
    },
    computed: {
        rowsPerPage: function() {
            if (this.pagerows == "all") return null;
            return parseInt(this.pagerows);
        },
        filteredRows: function() {
            return this.searchData(this.rows)
        },
    },
    methods: {
        resetStartRow: function() {
            this.startRow = 0;
        },
        searchData: function(data) {
            if (this.searchQuery.length == 0) {
                return data
            }
            return data.filter(obj => {
                for (var k of this.columns) {
                    const value = obj[k];
                    if (_.isString(value) && value.indexOf(this.searchQuery) >= 0) {
                        return true
                    }
                }
                return false
            })
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
                posty(this.dataURL + id, data, "PATCH").then(router.go(-1))
            } else {
                posty(this.dataURL, data).then(router.go(-1))
            }
        },
        deleteSelf: function() {
            posty(this.dataURL + this.myID(), null, "DELETE").then(this.showList())
        },
        showList: function(ev) {
            router.push(this.listURL)
        },
    }
}


var foundList = Vue.component("found-list", {
    template: "#tmpl-base-table",
    props: ["columns", "rows"],
    data: function () {
        return {
            sortKey: "",
            sortOrders: []
        }
    },
    methods: {
        /*
        results: function(data) {
        },
        */
        sortBy: function (key) {
            this.sortKey = key
            this.sortOrders[key] = this.sortOrders[key] * -1
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return foundLink(entry.Kind) + entry["ID"]
        }
    },
    watch: {
        "columns": function() {
            var sortOrders = {}
            this.columns.forEach(function (key) {
                sortOrders[key] = 1
            })
        }
    },
    events: {
        "found-these": function(funny) {
            this.rows = funny
        }
    }
})


// remove leading/trailing spaces, non-ascii
// TODO: perhaps just "printable" chars?
var cleanText = function(text) {
    text = text.replace(/^[^A-Za-z0-9:\.\-]*/g, "")
    text = text.replace(/[^A-Za-z0-9:\.\-]*$/g, "")
    text = text.replace(/[^A-Za-z 0-9:\.\-]*/g, "")
    return text
}

var searchFor = Vue.component("search-for", {
    template: "#tmpl-search-for",
    data: function () {
        return {
            columns: ["Kind", "Name", "Note"],
            searchText: "",
            found: [],
        }
    },
    created: function() {
        this.search()
    },
    methods: {
        search: function() {
            this.searchText = this.$route.params.searchText
            if (this.searchText.length === 0) {
                return
            }
            var url = searchURL + this.searchText;
            get(url).then(data => {
                if (data) {
                    if (data.length == 1) {
                        var link = foundLink(data[0].Kind) + data[0].ID
                        router.push(link)
                    } else {
                        this.found = data
                    }
                } else {
                    this.found = []
                }
            })
        }
    },
})

Vue.component("main-menu", {
    template: "#tmpl-main-menu",
    data: function() {
       return {
           searchText: "",
           debug: false,
       }
    },
    computed: {
        "debugAction": function() {
            return this.debug ? "Disable" : "Enable"
        },
    },
    methods: {
        doSearch: function() {
            var text = cleanText(this.searchText);
            if (text.length == 0) {
                return
            }
            router.push({name: "searcher", params: { searchText: text }})
        },
        "toggleDebug": function() {
            this.debug = ! this.debug
            urlDebug = this.debug
        },
        "userinfo": function() {
            const user = fromCookie();
            if (user) {
                store.commit("setUser", user)
            }
        },
    }
})


var ipList = Vue.component("ip-list", {
    template: "#tmpl-ip-list",
    mixins: [pagedCommon, commonListMIX],
    created: function(ev) {
        this.loadData()
        this.title = "IP Addresses in Use"
    },
    data: function() {
        return {
            filename: "iplist",
            rows: [],
            columns: [
                "Site",
                "Type",
                "Host",
                "Hostname",
                "IP",
                "Note"
            ],
            sortKey: "",
            sortOrders: [],
            Host: "",
            STI: 1,
            IPT: 0,
            searchQuery: "",
            sites: [],
            typelist: [],

            // TODO: kindlist should be populated from device_types
            hostlist: [
               "",
                "VM",
                "Server",
                "Switch",
            ],
        }
    },
    created () {
        this.loadData()
    },
    methods: {
          loadData: function () {
              var url = networkURL;
              if (this.STI > 0) {
                  url +=  "?STI=" + this.STI;
              }
              get(url).then(data => {
                  if (data) {
                      this.rows = data
                  }
              })
              getSiteLIST(true).then(s => this.sites = s)
              getIPTypes().then(t => {
                    t.unshift({IPT:0, Name: "All"})
                    this.typelist = t
              })
        },
        linkable: function(key) {
            return (key == "Hostname")
        },
        linkpath: function(entry, key) {
            if (entry.Host != "VM") {
                return "/device/edit/" + entry["ID"]
            }
            return "/vm/edit/" + entry["ID"]
        },
    },
    computed: {
        filteredRows: function() {
            var data;
            if (this.IPT || this.Host) {
                data = this.rows.filter(obj => {
                    if (this.IPT == obj.IPT && ! this.Host) {
                        return true
                    }
                    if (this.Host == obj.Host && ! this.IPT) {
                        return true
                    }
                    return (this.Host == obj.Host && this.IPT == obj.IPT)
                });
            } else {
                data = this.rows
            }
            return this.searchData(data)
        },
    },
    watch: {
        "STI": function(x) {
            this.loadData()
        }
    },
})

var reservedIPs = Vue.component("reserved-ips", {
    template: "#tmpl-reserved-ips",
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            filename: "iplist",
            rows: [],
            columns: [
                "Site",
                "VLAN",
                "IP",
                "Note",
                "User",
            ],
            sortKey: "",
            sortOrders: [],
            STI: 0,
            searchQuery: "",
            sites: [],
        }
    },
    created: function() {
        this.common()
    },
    methods: {
        common: function () {
            getSiteLIST(true).then(s => this.sites = s);
            get(reservedURL).then(r => this.rows = r || []);
        },
        linkable: function(key) {
            return (key == "IP")
        },
        linkpath: function(entry, key) {
            return "/ip/edit/" + entry["IID"]
        },
    },
    computed: {
        filteredRows: function() {
            var data;
            if (this.STI > 0) {
                data = this.rows.filter(obj => (this.STI == obj.STI))
            } else {
                data = this.rows
            }
            return this.searchData(data)
        },
    },
})


//
// IP TYPES
//
var ipTypes = Vue.component("ip-types", {
    template: "#tmpl-ip-types",
    mixins: [pagedCommon],
    data: function() {
        return {
            rows: [],
            columns: [
                "Name",
                "Multi"
            ],
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            getIPTypes().then(t => this.rows = t)
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return "/ip/type/edit/" + entry["IPT"]
        }
    },
    watch: {
        "STI": function(x) {
            this.loadData()
        }
    },
})


var ipTypeEdit = Vue.component("iptype-edit", {
    template: "#tmpl-iptype-edit",
    mixins: [authVue],
    data: function() {
        return {
            IPType: {}
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            if (this.$route.params.IPT > 0) {
                getIPTypes(this.$route.params.IPT).then(t => this.IPType = t)
            } else {
                this.IPType = {IPT: 0}
            }
        },
        newname: function() {
            console.log("my name is:", this.IPType.Name)
        },
        saveSelf: function() {
            var data = this.IPType;
            var id = data.IPT;
            var url = iptypesURL;
            if (id > 0) {
                posty(url + id, data, "PATCH").then(x => this.showList())
            } else {
                posty(url + id, data).then(x => this.showList())
            }
        },
        deleteSelf: function() {
        },
        showList: function() {
            router.push("/ip/types")
        },
    },
})


//
// User List
//

var userList = Vue.component("user-list", {
    template: "#tmpl-user-list",
    mixins: [pagedCommon],
    data: function() {
        return {
            columns: ["Login", "First", "Last", "Level"],
            rows: [],
            url: userURL,
            searchQuery: "",
        }
    },
    created: function() {
        this.loadData()
    },
    methods: {
        loadData: function() {
            get(this.url).then(data => this.rows = data || [])
        },
        linkable: function(key) {
            return (key == "Login")
        },
        linkpath: function(entry, key) {
            return "/user/edit/" + entry["USR"]
        }
    }
})

//
// USER EDIT
//

var userEdit = Vue.component("user-edit", {
    template: "#tmpl-user-edit",
    mixins: [editVue],
    data: function() {
        return {
            User: {},
            dataURL: userURL,
            listURL: "/user/list",

            // TODO: pull levels data from server
            levels: [
                {Level:0, Label: "User"},
                {Level:1, Label: "Editor"},
                {Level:2, Label: "Admin"},
            ],
        }
    },
    created: function() {
        this.loadSelf()
    },
    computed: {
        canAssume: function() {
            return this.$store.getters.isAdmin
        },
        disableAdd: function() {
            return (!(this.User.First && this.User.Last && this.User.Login && this.User.Level == 0))
        },
        showKey: function() {
            return ((this.User.USR && this.$store.getters.isAdmin) || (this.$store.getters.USR == this.User.USR));
        },
    },
    methods: {
        loadSelf: function () {
            if (this.$route.params.USR > 0) {
                getUser(this.$route.params.USR).then(u => this.User = u)
            } else {
                this.User = {USR: 0, Level: 0}
            }
        },
        myID: function() {
            return this.User.USR
        },
        myself: function() {
            return this.User
        },
        assumeUser: function() {
            var url = assumeURL + this.User.USR;
            posty(url, null).then(user => {
                this.$dispatch("user-auth", user)
                router.push("/")
            })
        },
    },
})



var ipMIX = {
    props: ["iptypes", "ports"],
    data: function() {
        return {
            newIP: "",
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
            posty(ipURL + iid, data, "PATCH")
            return false
        },
        deleteIP(i) {
            var iid = this.rows[i].IID
            posty(ipURL + iid, null, "DELETE").then(() => this.rows.splice(i, 1))
        },
        addIP: function() {
            var data = {IFD: this.newIFD, IPT: this.newIPT, IP: this.newIP}
            posty(ipURL, data).then(ip => {
                this.rows.push(ip)
                this.newIP = ""
                this.newIPT = 0
                this.newIFD = 0
            })
            return false
        }
    }
}

var netgrid = childTable("network-grid", "#tmpl-network-grid", [ipMIX])

var interfaceMIX = {
    props: ["DID", "IPMI"],
    data: function() {
        return {
            newPort: "",
            newMgmt: false,
            newMAC: "",
            newSwitchPort: "",
            newCableTag: "",
        }
    },
    computed: {
        interfaceAddDisabled: function() {
            return this.newPort.length == 0 || this.newMAC.length == 0
        }
    },
    methods: {
        findMAC(i) {
            get(deviceMacURL + this.DID).then(mac => {
                var row = this.rows[i]
                row.MAC = mac.MAC
                var data = {
                    IFD: row.IFD,
                    MAC: row.MAC,
                }
                posty(ifaceURL + row.IFD, data, "PATCH")
            })
        },
        needsMAC(i) {
            var iface = this.rows[i];
            return (iface.Port == "Eth0" && iface.MAC.length == 0 && ! iface.Mgmt)
        },
        updateInterface(i) {
            var row = this.rows[i]
            var ifd = row.IFD

            var port = row.Port.replace(/[^\d]*/g, "");
            port = (port.length) ? parseInt(port) : 0

            var data = {
                IFD: ifd,
                Port: port,
                Mgmt: row.Mgmt,
                MAC: row.MAC,
                SwitchPort: row.SwitchPort,
                CableTag: row.CableTag,
            }
            posty(ifaceURL + ifd, data, "PATCH")
        },
        deleteInterface(i) {
            var ifd = this.rows[i].IFD
            posty(ifaceURL + ifd, null, "DELETE")
                .then(() => this.rows.splice(i, 1))
                .catch(ack => console.log("============= ACK!:", ack))
        },
        addInterface: function() {
            var port = this.newPort.replace(/[^\d]*/g, "");
            port = (port.length) ? parseInt(port) : 0
            var data = {
                did: this.DID,
                port: port,
                mgmt: this.newMgmt,
                mac: this.newMAC,
                switchport: this.newSwitchPort,
                cabletag: this.newCableTag,
            }
            posty(ifaceURL, data).then(iface => {
                if (! iface.Mgmt) {
                    iface.Port = "Eth" + iface.Port
                }
                this.rows.push(iface)

                this.newPort = ""
                this.newMgmt = ""
                this.newMAC = ""
                this.newSwitchPort = ""
                this.newCableTag = ""
            })
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
            adding: false,
            sites: [],
            device_types: [],
            mfgrs: [],
            vmFilename: "vms",
            tags: [],
            ipTypes: [],
            newIP: "",
            IPMI: "huh",
            newIPD: 0,
            newIFD: 0,
            netColumns: ["IP", "Type", "Port"],
            ifaceColumns: ["Port", "Mgmt", "MAC", "CableTag", "SwitchPort"],
            part_columns: [
                "Description",
                "PartSN",
                "PartNumber",
                "Vendor",
            ],
            Description: "",
            Device: {},
            pageRows: 10,
            startRow: 0,
            newVM: false,
            vmColumns: ["Hostname", "IPs", "Profile", "Note"],
            parts: [],
        }
    },
    computed: {
        urlSuffix: function() {
            return this.Device.DC + "/" + this.Device.RID
        },
        disableSave: function() {
            const hostname  = this.Device.Hostname || "";
            return (hostname.length == 0 
                    || (! this.Device.STI > 0) 
                    || (! this.Device.DTI > 0) 
                    || (! this.Device.RID > 0)
                    || (! this.Device.RU > 0)
                    || (! this.Device.Height > 0)
                   )
        },
        canAddVM: function() {
            return (this.Device.DID > 0 && ! this.newVM)
        }
    },
    created: function() {
        if (! this.adding) {
            this.loadDevice()
        }
        this.loadCommon()
    },
    methods: {
        loadCommon: function() {
            getSiteLIST(false).then(s => this.sites = s)
            getDeviceTypes().then(t => this.device_types = t)
            getTagList().then(l => this.tags = l)
            getIPTypes().then(t => this.ipTypes = t)
            getMfgr().then(m => this.mfgrs = m)
        },
        loadDevice: function() {
            if (this.$route.params.DID > 0) {
                getDevice(this.$route.params.DID)
                    .then(getInterfaces)
                    .then(deviceRacks)
                    .then(deviceVMlist)
                    .then(d => {
                        this.Device = d
                        for (var ip of d.ips) {
                            if (ip.Mgmt) {
                                this.IPMI = ip.IP
                                break;
                            }
                        }
                    })
                getPart(this.$route.params.DID, "?did=").then(p => this.parts = p)
            } else {
                this.Device = {
                    DID: 0,
                    DTI: null,
                    TID: null,
                    MID: null,
                    Height: 1,
                    STI: null,
                    RID: null,
                    RU: null,
                }
            }
        },
        loadRacks: function() {
            const url = rackViewURL + "?STI=" + this.Device.STI;
            get(url).then(r => this.Device.racks = r)
        },
        saveSelf: function(event) {
            var device = this.Device;
            delete device["racks"];
            delete device["interfaces"];
            delete device["ips"];

            if (device.DID == 0) {
                posty(deviceViewURL, device).then(router.go(-1))
            } else {
                const url = deviceViewURL + this.Device.DID;
                posty(url, device, "PATCH").then(router.go(-1))
            }
        },
        deleteSelf: function(event) {
            if (window.confirm("Really delete this device?")) {
                const url = deviceViewURL + this.Device.DID;
                posty(url, null, "DELETE").then(router.go(-1))
            }
        },
        showList: function(ev) {
            router.go(-1)
        },
        portLabel: function(ipinfo) {
            if (this.Device.Type === "server") {
                return ipinfo.Mgmt ? "IPMI" : "Eth" + ipinfo.Port
            }
            return (ipinfo.Mgmt ? "Mgmt" : "Port") + ipinfo.Port
        },
        getMacAddr: function(ev) {
            const url = macURL + this.Server.IPIpmi;
            get(url).then(data => this.Device.MacPort0 = data.MacEth0)
        },
        vmLinkable: function(key) {
            return (key == "Hostname")
        },
        vmLinkpath: function(entry, key) {
            if (key == "Hostname") return "/vm/edit/" + entry["VMI"]
        },
        addVM: function() {
            this.newVM = true;
        },
        partLinkable: function(key) {
            return (key == "Description")
        },
        partLinkpath: function(entry, key) {
            if (key == "Description") return "/part/edit/" + entry["PID"]
        },
        audit: function() {
            router.push("/device/audit/" + this.Device.DID)
        }
    },
    watch: {
        "Device.STI": {
            handler: function (val, oldVal) {
                console.log("watch STI:", this.Device.STI);
                this.loadRacks()
            },
            deep: true
        },
        "$route": "loadDevice"
    }
}

var deviceEdit = Vue.component("device-edit", {
    template: "#tmpl-device-edit",
    mixins: [deviceEditVue],
})



var deviceAdd = Vue.component("device-add", {
    template: "#tmpl-device-edit",
    mixins: [deviceEditVue],
    created: function() {
        this.newDevice()
    },
    data: function() {
        return {
            adding: true
        }
    },
    methods: {
        newDevice: function() {
            var device = {
                DID: 0,
                DTI: null,
                TID: null,
                MID: null,
                Height: 1,
                STI: parseInt(this.$route.params.STI),
                RID: parseInt(this.$route.params.RID),
                RU: parseInt(this.$route.params.RU),
            }
            deviceRacks(device).then(d => this.Device = d)
        },
    },
})

var deviceVMs = childTable("device-vms", "#tmpl-base-table")

//
// DEVICE LOAD
//
//
var deviceLoad = Vue.component("device-load", {
    template: "#tmpl-device-load",
    data: function() {
        return {
            Devices: "",
            racks: [],
            sites: [],
            STI: 0,
        }
    },
    created: function() {
        this.loadCommon()
    },
    methods: {
        loadCommon: function () {
            getSiteLIST().then(s => this.sites = s)
            if (this.STI > 0) {
                getRack(this.STI, "?sti=").then(r => this.racks = r)
            } else {
                this.racks = []
            }
        },
        showList: function(ev) {
            router.push("/device/list/" + this.STI)
        },
        // TODO: copied from device edit -- merge functionality?
        addInterface: function(DID, Port, Mgmt, MAC, SwitchPort, CableTag, callBack) {
            Port = port.replace(/[^\d]*/g, "");
            Port = (port.length) ? parseInt(Port) : 0
            var data = {
                DID: DID,
                Port: Port,
                Mgmt: Mgmt,
                MAC: MAC,
                SwitchPort: SwitchPort,
                CableTag: CableTag,
            }
            if (callBack) {
                posty(ifaceURL, data).then(callBack)
            } else {
                posty(ifaceURL, data)
            }
        },
        addNetwork: function(DID, Port, Mgmt, MAC, SwitchPort, CableTag, IP) {
            Port = Port.replace(/[^\d]*/g, "");
            Port = (Port.length) ? parseInt(Port) : 0
            var data = {
                DID: DID,
                Port: Port,
                Mgmt: Mgmt,
                MAC: MAC,
                SwitchPort: SwitchPort,
                CableTag: CableTag,
            }
            posty(ifaceURL, data).then(function(iface) {
                var ipinfo = {IFD: iface.IFD, IP: IP}
                posty(ipURL, ipinfo)
            })
        },
        saveDevices: function() {
            // normalize column names
            var normalize = function(col) {
                switch (col) {
                    case "ip address":      return "IP";
                    case "ipmi address":    return "IPMI";
                    case "ipmi cable":      return "IPMICable";
                    case "ipmi port":       return "IPMIPort";
                    case "asset-tag":       return "AssetTag";
                    case "rack":            return "Rack";
                    case "ru":              return "RU";
                    case "hostname":        return "Hostname";
                    case "height":          return "Height";
                    case "alias":           return "Alias";
                    case "profile":         return "Profile";
                    case "note":            return "Note";
                }
            }
            var lines = this.Devices.split("\n");
            var cols = {};      // Device specific columns
            var lookup = {};    // All columns
            var intVals = ["Height", "Rack", "RU", "Version"];

            // parse column headers
            var line = lines[0].split("\t");
            for (var k=0; k < line.length; k++) {
                var head = line[k].toLowerCase();
                var col = normalize(head)
                if (col) {
                    cols[k] = col
                }
                lookup[head] = k
            }

            for (var i=1; i < lines.length; i++) {
                if (lines[i].trim().length == 0) {
                    continue
                }
                const line = lines[i].split("\t");
                if (line.length < 1) {
                    continue
                }

                var device = {
                    DID: null,
                    RID: null,
                    MID: null,
                    DTI: null,
                    TID: null,
                    STI: this.STI,
                }
                for (var j in cols) {
                    var col = cols[j];
                    if (intVals.indexOf(col) > -1) {
                        device[col] = parseInt(line[j]);
                    } else {
                        device[col] = line[j];
                    }
                }
                for (var r in this.racks) {
                    if (device.Rack == this.racks[r].Label) {
                        device.RID = this.racks[r].RID
                        break
                    }
                }
                var url = deviceURL;
                posty(url, device).then(added => {
                    console.log("added hostname:",  added.Hostname, " DID:", added.DID)
                    var ip    = line[lookup["ip_ipmi"]];
                    var port  = line[lookup["port_ipmi"]];
                    var mac   = line[lookup["mac_ipmi"]];
                    var cable = line[lookup["cable_ipmi"]];
                    this.addNetwork(added.DID, port, true, mac, port, cable, ip)

                    var ip    = line[lookup["ip_internal"]];
                    var port  = line[lookup["port_eth0"]];
                    var mac   = line[lookup["mac_eth0"]];
                    var cable = line[lookup["cable_eth0"]];
                    this.addNetwork(added.DID, port, false, mac, port, cable, ip)
                })
            }
            this.showList()
        },
    },
    watch: {
        "STI": function() {
            racks: getRack(this.STI, "?sti=")
       }
   }
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
                    for (var keep of ignore) {
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
var deviceAudit = Vue.component("device-audit", {
    template: "#tmpl-device-audit",
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
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            const ignore = ["TS", "Version", "Login", "USR", "RID", "TID", "KID", "DTI"];
            getDeviceAudit(this.$route.params.DID).then(fix => rows = deltas(ignore, fix))
        },
        linkable: function(key) {
        },
        linkpath: function(entry, key) {
        }
    },
    watch: {
        "STI": function(x) {
            this.loadData()
        }
    },
})


//
// VM IPs
//
var vmips = Vue.component("vm-ips", {
    template: "#tmpl-vm-ips",
    props: ["VMI"],
    data: function() {
        return {
            newIP: "",
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
            const url = ipURL + "?VMI=" + this.VMI;
            get(url).then(data => this.rows = data || [])
            get(iptypesURL).then(data => this.types = data)
        },
        updateIP(i) {
            var row = this.rows[i]
            var iid = row.IID
            var ip = row.IP
            var ipt = row.IPT
            var data = {VMI:this.VMI, IID: iid, IPT: ipt, IP: ip}
            posty(ipURL + iid, data, "PATCH")
        },
        deleteIP(i) {
            var iid = this.rows[i].IID
            posty(ipURL + iid, null, "DELETE")
                .then(() => this.rows.splice(i, 1))
        },
        addIP: function() {
            var data = {VMI: this.VMI, IPT: this.newIPT, IP: this.newIP}
            posty(ipURL, data).then(ips => {
                this.rows.push(data)
                this.newIP = ""
                this.newIPT = 0
            })
        }
    },
    watch: {
        "VMI": function() {
            this.loadSelf()
        }
    }
})


//
// VM Edit
//
var vmEdit = Vue.component("vm-edit", {
    template: "#tmpl-vm-edit",
    mixins: [siteMIX],
    data: function() {
        return {
            url: vmViewURL,
            STI: 0,
            sites: [],
            racks: [],
            tags: [],
            ipTypes: [],
            ipRows: [],
            Description: "",
            VMI: parseInt(this.$route.params.VMI),
            VM: {VMI: null, Server: ''},
        }
    },
    created: function() {
       this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            if (this.$route.params.VMI > 0) {
                getVM(this.$route.params.VMI).then(v => this.VM = v).catch(meh => {
                    this.showList()
                })
            } else {
                this.VM = {VMI: null, Server: ''}
            }
        },
        saveSelf: function() {
            posty(this.url + this.VM.VMI, this.VM, "PATCH").then(() => this.showList)
        },
        deleteSelf: function() {
            if (window.confirm("Really delete this VM?")) {
                posty(this.url + this.VM.VMI, null, "DELETE").then(() => this.showList)
            }
        },
        showList: function(ev) {
            router.push("/vm/list")
        },
    },
})

var vmAdd = Vue.component("vm-add", {
    template: "#tmpl-vm-edit",
    props: ["DID", "Server"],
    data: function() {
        return {
            url: vmViewURL,
            tags: [],
            ipTypes: [],
            ipRows: [],
            Description: "",
            VM: {VMI:null, DID: 0, Server:''}
        }
    },
    created: function() {
        console.log("THIS SERVER:", this.Server)
        this.VM['DID'] = this.DID
        this.VM['Server'] = this.Server
    },
    methods: {
        saveSelf: function() {
            if (this.VM.VMI > 0) {
                posty(this.url + this.VM.VMI, this.VM, "PATCH").then(() => router.go(-1))
            } else {
                posty(vmURL, this.VM).then(vm => {
                    this.$parent.Device.vms.push(vm)
                    this.$parent.newVM = false
                })
            }
        },
        deleteSelf: function() {
            if (window.confirm("Really delete this VM?")) {
                posty(this.url + this.VM.VMI, null, "DELETE").then(() => this.showList)
            }
        },
        showList: function(ev) {
            router.go(-1)
            //router.push("/vm/list")
        },
    },
})


// audit data comes back with column and row data separate
// this will create our standard row with named fields
var vmAudit = Vue.component("vm-audit", {
    template: "#tmpl-audit",
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
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            getVMAudit(this.$route.params.VMI, "?vmi=").then(fix => {
                const ignore = ["TS", "Version", "Login", "USR", "RID", "TID", "KID", "DTI"];
                this.rows = deltas(ignore, fix)
            })
        },
        linkable: function(key) {
        },
        linkpath: function(entry, key) {
        }
    },
})


//
// DEVICE LIST
//

var deviceList = Vue.component("device-list", {
    template: "#tmpl-device-list",
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            STI: 1,
            RID: 0,
            DTI: 0,
            sites: [],
            racks: [],
            searchQuery: "",
            rows: [],
            filename: "servers",
            types: [],
            columns: [
                 "Site",
                 "Rack",
                 "RU",
                 "Hostname",
                 "Alias",
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
    computed: {
        filteredRows: function() {
            var data = this.searchData(this.rows);

            if (this.STI > 0 && this.RID > 0) {
                data = data.filter(obj => (obj.RID == this.RID))
            }
            if (this.DTI > 0) {
                data =  data.filter(obj => (obj.DTI == this.DTI))
            }
            return data
        }
    },
    created () {
        this.loadDevices()
        this.loadData()
    },
    methods: {
        loadDevices: function() {
            if (this.STI > 0) {
                getDeviceLIST(this.STI, "?sti=").then(devices => this.rows = devices || [])
                getRack(this.STI, "?sti=").then(r => {
                    if (this.STI > 0) r.unshift({RID:null, Label: ""})
                    this.racks = r || []
                })
            } else {
                getDeviceLIST().then(d => this.rows = d || [])
                this.racks = []
            }
        },
        loadData: function() {
            this.RID = 0;
            getSiteLIST(true).then(s => this.sites = s)
            getDeviceTypes().then(t => this.types = t)
        },
        canLink: function(column) {
            return column === "Hostname"
        },
        linkFN: function(entry, key) {
            if (key == "Hostname") return "/device/edit/" + entry["DID"]
        }
    },
    watch: {
        "STI": function(val, oldVal) {
            mySTI = val
            this.loadDevices()
        },
    },
})


//
// DEVICE TYPES
//

var deviceTypes = Vue.component("device-types", {
    template: "#tmpl-device-types",
    mixins: [pagedCommon, commonListMIX],
    data: function() {
      return {
          searchQuery: "",
          rows: [],
          columns: [
               "Name",
            ],
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            getDeviceTypes().then(t => this.rows = t)
        },
        addType: function() {
            router.push("/device/type/edit/0")
        },
        linkable: function(column) {
            return column === "Name"
        },
        linkpath: function(entry, key) {
            if (key == "Name") return "/device/type/edit/" + entry["DTI"]
        }
    },
})

var deviceTypeEdit = Vue.component("device-type-edit", {
    template: "#tmpl-device-type-edit",
    data: function() {
        return {
            DeviceType: {}
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
              if (this.$route.params.DTI > 0) {
                  getDeviceTypes(this.$route.params.DTI).then(dt => this.DeviceType = dt)
              } else {
                  this.DeviceType = {DTI: 0}
              }
        },
        saveSelf: function() {
            var url = deviceTypesURL;
            if (this.DeviceType.DTI > 0) {
                url += this.DeviceType.DTI
                posty(url, this.DeviceType, "PATCH").then(() => this.showList())
            } else {
                posty(url, this.DeviceType).then(() => this.showList())
            }
        },
        showList: function() {
            router.push("/device/types")
        },
        deleteSelf: function() {
            var url = deviceTypesURL + this.DeviceType.DTI
            posty(url, null, "DELETE").then(this.showList);
        },
    },
})


//
// VLAN List
//

var vlanList = Vue.component("vlan-list", {
    template: "#tmpl-vlan-list",
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            filename: "vlans",
            dataURL: vlanViewURL,
            listURL: "/vlan/list",
            STI: 0,
            sites: [],
            searchQuery: "",
            rows: [],
            columns: [
                "Site",
                "Name",
                "Profile",
                "Netmask",
                "Gateway",
                "Route",
                "Starting",
                "Note",
            ]
         }
    },
    created: function () {
        this.loadSelf()
        getSiteLIST(true).then(s => this.sites = s)
    },
    methods: {
        loadSelf: function () {
            const url = this.dataURL;
            get(url).then(data => this.rows = data)
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return "/vlan/edit/" + entry["VLI"]
        },
    },
    computed: {
        filteredRows: function() {
            if (this.STI > 0) {
                return this.rows.filter(obj => (obj.STI == this.STI))
            }
            return this.rows
        }
    }
})


//
// VLAN Edit
//

var vlanEdit = Vue.component("vlan-edit", {
    template: "#tmpl-vlan-edit",
    mixins: [editVue, commonListMIX, siteMIX],
    data: function() {
        return {
            dataURL: vlanURL,
            sites: [],
            VLAN: {},
            STI: 0,
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf() {
            getVLAN(this.$route.params.VLI).then(v => {
                this.VLAN = v
                this.STI = this.VLAN.STI
            })
            getSiteLIST(true).then(s => this.sites = s)
        },
        myID: function() {
              return this.VLAN.VLI
        },
        myself: function() {
              return this.VLAN
        },
        showList: function(ev) {
              router.push("/vlan/list")
        },
    },
})


//
// IP Reserve
//
var ipReserve = Vue.component("ip-reserve", {
    template: "#tmpl-ip-reserve",
    data: function() {
        return {
            conflicted: 0,
            sites: [],
            vlans: [],
            From: "",
            To: "",
            Note: "",
            Network: "",
            Netmask: "",
            Max: "",
            STI: 0,
            VLI: 0,
            minIP32: 0,
            maxIP32: 0,
            VLAN: {},
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
    created: function() {
        this.common()
    },
    methods: {
        common: function() {
            getSiteLIST().then(s => this.sites = s);
        },
        showList: function() {
            router.push("/ip/reserved")
        },
        reserveIPs: function() {
            var url = ipReserveURL;
            var data = {
                From: this.From,
                To: this.To,
                VLI: this.VLI,
                Note: this.Note,
            }
            posty(url, data).then(this.showList).catch(fail => {
                var count=fail["Count"];
                if (count > 0) {
                    this.conflicted = count;
                }
            })
        },
        checkFrom: function() {
            if (! validIP(this.From)) {
                alert("Invalid IP:", this.From)
            }
        },
        checkTo: function() {
            if (! validIP(this.To)) {
                alert("Invalid IP:", this.To)
            }
        },
    },
    watch: {
        STI: function() {
            const url = vlanURL + "?STI=" + this.STI
            get(url).then(data => this.vlans = data.sort((a, b) => a.Name > b.Name))
        },
        VLI: function() {
            for (var i=0; i < this.vlans.length; i++) {
                var vlan = this.vlans[i]
                if (vlan.VLI == this.VLI) {
                    var mask   = ip32(vlan.Netmask)
                    var net    = ip32(vlan.Gateway)
                    var min_ip = (net & mask) + 1
                    var max_ip = (net | ~mask) - 1

                    this.minIP32 = min_ip
                    this.maxIP32 = max_ip
                    this.Network = ipv4(min_ip) + " - " + ipv4(max_ip)
                    break
                }
            }
        }
    }
})


//
// IP EDIT
//
var ipEdit = Vue.component("ip-edit", {
    template: "#tmpl-ip-edit",
    mixins: [authVue],
    data: function() {
        return {
            IP: {},
            iptypes: [],
        }
    },
    computed: {
        "inuse": function() {
            return ((this.IP.VMI > 0) || (this.IP.IFD > 0))
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            const url = ipViewURL + this.$route.params.IID;
            get(url).then(ip => this.IP = ip)
            get(iptypesURL).then(t => this.iptypes = t)
        },
        showList: function() {
            router.push("/ip/list")
        },
        saveSelf: function(event) {
            var url = ipURL + this.IP.IID
            posty(url, this.IP, "PATCH").then(this.showList)
        },
        deleteSelf: function() {
            var url = ipURL + this.IP.IID
            posty(url, this.IP, "DELETE").then(this.showList)
        },
    },
})


//
// VM LIST
//

var vmList = Vue.component("vm-list", {
    template: "#tmpl-vm-list",
    mixins: [pagedCommon, siteMIX, commonListMIX ],
    data: function() {
        return {
            filename: "vms",
            STI: 1,
            sites: [],
            site: "",
            searchQuery: "",
            rows: [],
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
            getSiteLIST(true).then(f => this.sites = f)
            var url = vmIPsURL;
            if (this.STI > 0) {
                url += "?sti=" + this.STI
            }
            get(url).then(data => this.rows = data || [])
        },
        linkable: function(key) {
            return (key == "Hostname" || key == "Server")
        },
        linkpath: function(entry, key) {
            if (key == "Server")   return "/device/edit/" + entry["DID"]
            if (key == "Hostname") return "/vm/edit/" + entry["VMI"]
        }
  },
  watch: {
    "STI": function(val, oldVal){
            this.loadSelf()
        },
    },
})


//
// Inventory
//

var partInventory = Vue.component("part-inventory", {
    template: "#tmpl-part-inventory",
    mixins: [pagedCommon, commonListMIX, siteMIX],
    data: function() {
        return {
            DID: 0,
            STI: 4,
            PID: 0,
            available: [],
            sites: [],
            hostname: "",
            other: "",
            searchQuery: "",
            rows: [],
            columns: ["Site", "Description", "PartNumber", "Mfgr", "Qty", "Price"],
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            if (this.STI > 0) {
                getInventory(this.STI, "?sti=").then(r => this.rows = r || [])
            } else {
                getInventory().then(r => this.rows = r || [])
            }
            getSiteLIST(true).then(s => this.sites = s)
        },
        updated: function(event) {
            console.log("the event: " + event)
        },
        linkable: function(key) {
            return (key == "Description")
        },
        linkpath: function(entry, key) {
            return "/part/use/" + entry["STI"] + "/" + entry["KID"]
        },
    },
    watch: {
        "STI": function(val, oldVal){
            this.loadSelf()
        },
    },
})


var partUse = Vue.component("part-use", {
    template: "#tmpl-part-use",
    data: function() {
        return {
            badHost: false,
            DID: 0,
            STI: 4,
            PID: 0,
            available: [],
            sites: [],
            hostname: "",
            other: "",
            searchQuery: "",
            partData: [],
        }
    },
    computed: {
        "unusable": function() {
            return (((this.hostname.length == 0) || this.badHost) && this.other.length == 0);
        }
    },
    created: function () {
        this.loadData()
    },
    methods: {
        showList: function() {
            router.push("/part/inventory")
        },
        loadData: function () {
            var kid = this.$route.params.KID;
            var sti = this.$route.params.STI;
            var url = partURL + "?unused=1&bad=0&kid=" + kid + "&sti=" + sti
            get(url).then(data => this.available = data)
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
            var url = deviceViewURL + "?hostname=" + this.hostname;
            get(url).then(hosts => {
                if (hosts && hosts.length == 1) {
                    this.DID = hosts[0].DID
                    this.badHost = false;
                } else {
                    this.badHost = true;
                }
            })
        },
    },
})


//
// Parts
//

var partList = Vue.component("part-list", {
    template: "#tmpl-part-list",
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
            hostname: "",
            other: "",
            searchQuery: "",
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
    created: function() {
        this.loadData()
    },
    methods: {
        fixPrices: function(parts) {
            for (var i=0; i<parts.length; i++) {
                parts[i].Price = parts[i].Price.toFixed(2);
            }
        },
        setRows: function(data) {
            this.rows = data || []
        },
        loadData: function() {
            this.STI = parseInt(this.$route.params.STI);
            getSiteLIST(true).then(s => this.sites = s)
            if (this.STI > 0) {
                //getPart(STI, "?sti=").then(p => this.rows = p.map(x => x.Price = Price.toFixed(2)))
                //getPart(STI, "?sti=").then(p => this.rows = (p || []))
                getPart(this.STI, "?sti=").then(p => this.setRows(p))
            } else {
                //getPart().then(p => this.rows = p.map(x => x.Price = Price.toFixed(2)))
                getPart().then(p => this.setRows(p))
            }
      },
      findhost: function(ev) {
          get("api/server/hostname/" + this.hostname).then((data, status) => {
               const enable = (status == 200);
               buttonEnable(document.getElementById("use-btn"), enable)
               this.DID = enable ? data.ID : 0;
            })
        },
        newPart: function(ev) {
            var id = parseInt(ev.target.id.split("-")[1]);
        },
        linkable: function(key) {
            return (key == "Description")
        },
        linkpath: function(entry, key) {
            return "/part/edit/" + entry["PID"]
        }
    },
    watch: {
        "STI": function(newVal,oldVal) {
            router.push("/part/list/" + newVal)
        },
        $route () {
            this.loadData()
        }
    },
    computed: {
        filteredParts: function() {
            var data = this.searchData(this.rows)
            if (this.ktype == 1) {
                return data
            }
            return data.filter(obj => {
                if (this.ktype == 2 && ! obj.Bad) {
                    return true
                }
                return (this.ktype == 3 && obj.Bad)
            });
        },
    }
})


//
// PART TYPES
//

var partTypes = Vue.component("part-types", {
    template: "#tmpl-part-types",
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            searchQuery: "",
            rows: [],
            columns: [
               "Name",
            ]
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            getPartTypes().then(p => this.rows = p)
        },
        addType: function() {
            router.push("/part/type/edit/0")
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return "/part/type/edit/" + entry["PTI"]
        }
    }
})


var partTypeEdit = Vue.component("part-type-edit", {
    template: "#tmpl-part-type-edit",
    mixins: [editVue],
    data: function() {
        return {
            PartType: {PTI: 0},
            dataURL: partTypesURL,
            listURL: "/part/type/list",
        }
    },
    created: function() {
        this.loadSelf()
    },
    computed: {
        notReady: function() {
            return (this.PartType.length == 0)
        }
    },
    methods: {
        loadSelf: function () {
            if (this.$route.params.PTI > 0) {
                getPartTypes(this.$route.params.PTI).then(t => this.PartType = t)
            } else {
                this.PartType = {PTI: 0, Name: ""}
            }
        },
        myID: function() {
            return this.PartType.PTI
        },
        myself: function() {
            return this.PartType
        },
        showList: function() {
            router.push(this.listURL)
        },
    },
})


//
// RMAs
//

var rmaList = Vue.component("rma-list", {
    template: "#tmpl-rma-list",
    mixins: [pagedCommon, commonListMIX],
    data: function() {
        return {
            STI: 0,
            sites: [],
            rmas: [],
            searchQuery: "",
            rmaType: 1,
            columns: [
                "RMD",
                "Site",
                "Description",
                "Hostname",
                "PartSN",
                "Vendor",
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
             var url = rmaviewURL;
             if (this.STI > 0) {
                 url += "?STI=" + this.STI
             }
             get(url).then(data => this.rmas = data || [])
             getSiteLIST(true).then(s => this.sites = s)
        },
        linkable: function(key) {
            switch(key) {
                case "Description": return true;
                case "Hostname": return true;
                case "RMD": return true;
            }
            return false;
        },
        linkpath: function(entry, key) {
            switch(key) {
                case "RMD": return "/rma/edit/" + entry["RMD"]
                case "Description": return "/part/edit/" + entry["OldPID"]
                case "Hostname":
                    if (!("DID" in entry)) return "";
                    return "/device/edit/" + entry["DID"]
            }
        },
    },
    computed: {
        rmaFiltered: function() {
            if (this.rmaType == 1) {
                return this.rmas
            }
            return this.rmas.filter(obj => {
                if (this.rmaType == 2 && ! obj.Closed) {
                    return true
                }
                return (this.rmaType == 3 && obj.Closed)
            })
        },
    },
    watch: {
        "STI": function(val, oldVal){
                this.loadSelf()
            },
    },
})


//
// VENDOR LIST
//

var vendorList = Vue.component("vendor-list", {
    template: "#tmpl-vendor-list",
    mixins: [pagedCommon],
    data: function() {
        return {
            sites: [],
            searchQuery: "",
            rows: [],
            columns: [
                "Name",
                //"WWW",
                "Phone",
/*
                "Address",
                "City",
                "State",
                "Country",
                "Postal",
*/
                "Note",
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            get(vendorURL).then(data => this.rows = data)
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return "/vendor/edit/" + entry["VID"]
        },
    },
})


var vendorEdit = Vue.component("vendor-edit", {
    template: "#tmpl-vendor-edit",
    mixins: [editVue],
    data: function() {
        return {
            Vendor: {VID: 0}, //vendor,
            dataURL: vendorURL,
            listURL: "/vendor/list",
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            if (this.$route.params.VID > 0) {
                getVendor(this.$route.params.VID).then(v => this.Vendor = v)
                return
            }
            this.Vendor = {VID: 0} // vendor
        },
        myID: function() {
            return this.Vendor.VID
        },
        myself: function() {
            return this.Vendor
        },
        showList: function() {
            router.push("/vendor/list")
        },
    },
})


//
// MFGR LIST
//

var mfgrList = Vue.component("mfgr-list", {
    template: "#tmpl-mfgr-list",
    mixins: [pagedCommon],
    data: function() {
        return {
            sites: [],
            searchQuery: "",
            rows: [],
            columns: [
                "Name",
                "Note",
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            get(mfgrURL).then(data => this.rows = data)
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return "/mfgr/edit/" + entry["MID"]
        },
    },
})



var mfgrEdit = Vue.component("mfgr-edit", {
    template: "#tmpl-mfgr-edit",
    data: function() {
        return {
            Mfgr: {}
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            if (this.$route.params.MID > 0) {
                getMfgr(this.$route.params.MID).then(m => this.Mfgr = m)
            } else {
                this.Mfgr = {MID: 0}
            }
        },
        showList: function(xhr) {
            router.push("/mfgr/list")
        },
        deleteSelf: function() {
            deleteIt(mfgrURL + this.Mfgr.MID, this.showList)
        },
        saveSelf: function()  {
            saveMe(mfgrURL, this.Mfgr, this.Mfgr.MID, this.showList)
        }
    },
    watch: {
        // call again the method if the route changes
        "$route": "loadSelf"
    },
})


//
// PART EDIT
//
var partEdit = Vue.component("part-edit", {
    template: "#tmpl-part-edit",
    data: function() {
        return {
            badHost: false,
            sites: [],
            types: [],
            vendors: [],
            Part: {},
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
    created: function() {
        this.loadCommon()
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            if (this.$route.params.PID > 0) {
                getPart(this.$route.params.PID).then(p => this.Part = p)
            } else {
                this.Part = {
                    PID: 0,
                    DID: 0,
                    STI: 0,
                    PTI: 0,
                    VID: 0,
                    Bad: false,
                    Used: false,
                }
            }
        },
        loadCommon: function() {
            getSiteLIST().then(s => this.sites = s)
            getPartTypes().then(t => this.types = t)
            getVendor().then(list => {
                list.unshift({VID:0, Name:""})
                this.vendors = list
            })
        },
        showList: function(ev) {
            router.push("/part/list/" + this.Part.STI)
        },
        validprice: function() {
        },
        saveSelf: function(event) {
            this.Part.Price = parseFloat(this.Part.Price)
            this.Part.Cents = Math.round(this.Part.Price * 100)
            var url = partURL;
            if (this.Part.PID > 0) {
                url += this.Part.PID
                posty(url, this.Part, "PATCH").then(this.showList)
            } else {
                posty(url, this.Part).then(this.showList)
            }
        },
        doRMA: function(ev) {
            router.push("/rma/create/" + this.Part.PID)
        },
        findhost: function() {
            if (this.Part.Hostname.length === 0) {
                this.Part.DID = 0
                this.badHost = false
                return
            }
            var url = deviceViewURL + "?hostname=" + this.Part.Hostname;
            get(url).then(hosts => {
                if (hosts && hosts.length == 1) {
                    this.Part.DID = hosts[0].DID
                    this.badHost = false;
                } else {
                    this.badHost = true;
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
        return {
            badHost: false,
            dataURL: rmaviewURL,
            RMA: {
                RMD: 0,
                NewPID: 0,
                OldPID: 0,
                UID: 0,
            }
        }
    },
    methods: {
        saveSelf: function(event) {
            if (this.RMA.RMD > 0) {
                posty(rmaviewURL + this.RMA.RMD, this.RMA, "PATCH").then(this.showList)
            } else {
                posty(rmaviewURL, this.RMA).then(this.showList)
            }
        },
        showList: function() {
            router.push("/rma/list")
        },
        findhost: function() {
            if (this.RMA.Hostname.length === 0) {
                this.RMA.DID = 0
                this.badHost = false
                return
            }
            var url = deviceViewURL + "?hostname=" + this.RMA.Hostname;
            get(url).then(hosts => {
                if (hosts && hosts.length == 1) {
                    this.RMA.DID = hosts[0].DID
                    this.badHost = false;
                } else {
                    this.badHost = true;
                }
            })
        },
    },
}

//
// RMAs
//

var rmaEdit = Vue.component("rma-edit", {
    template: "#tmpl-rma-edit",
    mixins: [rmaCommon],
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            getRMA(this.$route.params.RMD).then(r => this.RMA = r)
        },
        deleteSelf: function(event) {
            var url = this.dataURL + this.myID();
            posty(url, null, "DELETE").then(this.showList)
        },
    },
})


//
// RMA CREATE
//

var rmaCreate = Vue.component("rma-create", {
    template: "#tmpl-rma-edit",
    mixins: [rmaCommon],
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            getPart(this.$route.params.PID).then(part => {
                const now = new Date();
                const dd = now.getDate();
                const mm = now.getMonth()+1; //January is 0!
                const yyyy = now.getFullYear();
                const today = yyyy + "/" + mm + "/" + dd;
                this.RMA = {
                    RMD: 0,
                    DID: part.DID,
                    VID: part.VID,
                    STI: part.STI,
                    OldPID: part.PID,
                    NewPID: 0,
                    Description: part.Description,
                    PartNumber: part.PartNumber,
                    Hostname: part.Hostname,
                    PartSN: part.Serial,
                    DeviceSN: part.DeviceSN,
                    Created: today,
                }
            })
        },
    },
})


//
// PART LOAD
//
//
var partLoad = Vue.component("part-load", {
    template: "#tmpl-part-load",
    data: function() {
        return {
            Parts: "",
            sites: [],
            STI: 2,
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            getSiteLIST().then(s => this.sites = s)
        },
        showList: function(ev) {
            router.push("/part/list/" + this.STI)
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
                    continue
                }

                var part = {
                    PID: null,
                    KID: null,
                    DID: null,
                    VID: null,
                    STI: this.STI,
                    Bad: false,
                    Unused: true,
                }
                for (var j in cols) {
                    var col = cols[j];
                    part[col] = line[j];
                }

                const qty = parseInt(part["Qty"]) || 1

                if (part.Price) {
                    part.Price = part.Price.replace(/[^0-9.]*/g,"")
                    part.Price = parseFloat(part.Price)
                    part.Cents = Math.round(part.Price * 100)
                } else {
                    part.Price = 0.0
                    part.Cents = 0
                }

                for (var j=0; j < qty; j++) {
                    posty(partURL, part)
                }
            }
            this.showList()
        },
    },
})


//
// TAGS
//

var tagEdit = Vue.component("tag-edit", {
    template: "#tmpl-tag-edit",
    data: function () {
        return {
            tags: [],
            url: tagURL,
            tag: {TID: 0}, //tag,
            sites: [],
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        showList: function() {
            router.push("/")
        },
        loadSelf: function () {
             getSiteLIST().then(s => this.sites = s)
             get(this.url).then(data => this.tags = data)
        },
        deleteSelf: function(ev) {
            if (! this.tag) {
                return
            }
            posty(this.url + this.tag.TID, null, function(data) {}, "DELETE")
            for (var i=0; i < this.tags.length; i++) {
                if (this.tags[i].TID == this.tag.TID) {
                    this.tags.splice(i, 1)
                    break
                }
            }
            this.tag = {TID: 0}
        },
        saveSelf: function() {
            if (this.tag.TID > 0) {
                posty(this.url + this.tag.TID, this.tag, "PATCH").then(() => {
                    for (var i=0; i < this.tags.length; i++) {
                        if (this.tags[i].TID == this.tag.TID) {
                            this.tags[i].Name = this.tag.Name
                            break
                        }
                    }
                })
            } else {
                posty(this.url, this.tag).then(t => {
                    this.tag = t
                    this.loadSelf()
                })
            }
        },
    },
    watch: {
        "tag.TID": function() {
            for (var i=0; i < this.tags.length; i++) {
                if (this.tags[i].TID == this.tag.TID) {
                    this.tag.Name = this.tags[i].Name
                    return
                }
            }
            this.tag.Name = ""
        }
    },
})


//
// RACK Edit
//
var rackEdit = Vue.component("rack-edit", {
    template: "#tmpl-rack-edit",
    mixins: [editVue, siteMIX],
    data: function() {
        return {
            sites: [],
            id: "RID",
            name: "Rack",
            Rack: {
                RID: 0,
                STI: 0,
            },
            dataURL: rackViewURL,
            listURL: "/rack/list",
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
    created: function() {
        this.loadSelf()
        getSiteLIST(false).then(s => this.sites = s)
    },
    methods: {
        loadSelf: function () {
            if (this.$route.params.RID > 0) {
                getRack(this.$route.params.RID).then(r => this.Rack = r)
            } else {
                this.Rack = {
                    RID: 0,
                    STI: 0,
                }
            }
        },
        deleteSelf: function() {
            posty(rackViewURL + this.myID(), null, "DELETE").then(router.go(-1))
        },
        showList: function() {
            router.push("/rack/list")
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

var rackList = Vue.component("rack-list", {
    template: "#tmpl-rack-list",
    mixins: [pagedCommon, siteMIX, commonListMIX],
    data: function() {
        return {
            dataURL: rackViewURL,
            STI: 0,
            sites: [],
            searchQuery: "",
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
             var url = this.dataURL;
             if (this.STI > 0) {
                 url += "?sti=" + this.STI
             }
             get(url).then(data => this.rows = data)
        },
        linkable: function(key) {
            return (key == "Label")
        },
        linkpath: function(entry, key) {
            if (key == "Label") return "/rack/edit/" + entry["RID"]
        },
    },
    watch: {
        "STI": function(val, oldVal){
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
            Hostname:"",
            Mgmt:"",
            IPs:"",
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
            these[k]["pingMgmt"] = ""
            these[k]["pingIP"] = ""
        }
        lumps.push({rack: rack, units: these, other: unracked[rack.RID]})
    }
    return lumps
}


// for rack-layout
var rackView = Vue.component("rack-view", {
    template: "#tmpl-rack-view",
    props: ["layouts", "RID", "audit"],
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
            device.newHostname  = ""
            device.Height    = 0
            device.Hostname  = ""
            device.DID       = 0
            device.RID       = 0
            device.Mgmt      = ""
            device.IPs       = ""
        },
        moveUp: function(lay) {
            var ru = lay.RU + 1;
            var url = adjustURL + lay.DID;
            var adjust = {DID: lay.DID, RID: lay.RID, RU: ru, Height: lay.Height};
            posty(url, adjust, "PUT").then(moved => {
                if (moved.RU == ru) {
                    if (lay.Height > 1) this.zero(lay.RU + lay.Height)
                    this.copy(lay, ru)
                    this.zero(lay.RU)
                    this.rusize(lay.RU, 1)
                }
            })
        },
        moveDown: function(lay) {
            var ru = lay.RU - 1;
            var url = adjustURL + lay.DID;
            var adjust = {DID: lay.DID, RID: lay.RID, RU: ru, Height: lay.Height};
            posty(url, adjust, "PUT").then(moved => {
                if (moved.RU == ru) {
                    this.copy(lay, ru)
                    if (lay.Height > 1) {
                        this.rusize(ru + lay.Height, 1)
                    } else {
                        this.rusize(lay.RU, 1)
                    }
                    this.zero(lay.RU)
                }
            })
        },
        rackheight: function(lay) {
            return "rackheight" + lay.Height;
        },
        // TODO make common, pass in field of interest
        changeIP: function(lay) {
            if (! validIP(lay.newIP.trim())) {
                lay.badIP = true;
                return
            }
            var url = ifaceViewURL;
            url += "?did=" + lay.DID + "&ipv4=" + lay.IPs;
            get(url).then(function(data) {
                // TODO add error handling
                var ipinfo = data[0]
                var ip = {IID: ipinfo.IID, IP: lay.newIP}
                posty(ipURL + ipinfo.IID, ip, "PATCH").then(function(updated) {
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
            url += "?did=" + lay.DID + "&ipv4=" + lay.Mgmt;
            get(url).then(function(data) {
                // TODO add error handling
                if (! data || data.length != 1) {
                    return
                }
                var ipinfo = data[0]
                var ip = {IID: ipinfo.IID, IP: lay.newMgmt}
                posty(ipURL + ipinfo.IID, ip, "PATCH").then(function(updated) {
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
            // verify that new hostname doesn"t already exist
            getDevice(lay.newHostname,"?hostname=").then(function(device) {
                if (! device) {
                    var newname = {DID: lay.DID, Hostname: lay.newHostname}
                    posty(deviceViewURL + lay.DID, newname, "PATCH").then(function(good) {
                        lay.badHostname = false;
                    }).catch(function(fail) {
                        console.log("rename patch fail:", fail)
                    })
                } else {
                    lay.badHostname = true;
                }
            }).catch(function(fail) {
                    console.log("rename fail:", fail)
            })
        },
        resize: function(lay) {
            if (lay.newHeight < 1) {
                lay.badHeight = true;
                return
            }
            const url = adjustURL + lay.DID;
            const newsize = {DID: lay.DID, RID: lay.RID, RU: lay.RU, Height: lay.newHeight}
            posty(url, newsize, "PUT").then(adjusted => {
                if (adjusted.Height == lay.Height) {
                    lay.badHeight = true;
                    return
                }
                if (lay.newHeight > lay.Height) {
                    const to = this.layouts.rack.RUs - lay.RU;
                    const from = to - lay.newHeight + 1;
                    for (var i=from; i < to; i++ ) {
                        this.layouts.units[i].Height = 0;
                    }
                } else if (lay.newHeight < lay.Height) {
                    const from = this.layouts.rack.RUs - lay.RU - lay.Height;
                    const to   = from + (lay.Height - lay.newHeight) + 1
                    for (var i=from; i < to; i++ ) {
                        this.layouts.units[i].Height = 1;
                    }
                }
                lay.Height = lay.newHeight
                lay.badHeight = false;
            })
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

var rackLayout = Vue.component("rack-layout", {
    template: "#tmpl-rack-layout",
    mixins: [commonListMIX],
    data: function() {
        return {
            dataURL: deviceListURL,
            STI: 0,
            RID: 0,
            sites: [],
            racks: [],
            site: "",
            audit: false,
            lumpy:[],
        }
    },

    created: function () {
        this.loadSelf()
    },
    computed: {
        filteredRacks: function() {
            if (this.RID == 0) {
                return this.lumpy
            }
            return this.lumpy.filter(obj => (this.RID == obj.rack.RID))
        },
    },
    methods: {
        loadSelf: function () {
            var url = this.dataURL;
            if (this.$route.params.STI > 0) {
                this.STI = this.$route.params.STI
            } else {
                this.RID = 0
                this.STI = 0
            }

             if (this.RID > 0) {
                 url += "?rid=" + this.RID
             } else if (this.STI > 0) {
                 url += "?sti=" + this.STI
             }

             getSiteLIST().then(s => this.sites = s)
             get(url).then(units => {
                 url = rackViewURL + "?STI=" + this.STI;
                 get(url).then(racks => {
                     if (racks) {
                         racks.unshift({RID:0, Label:""})
                         this.racks = racks
                         this.lumpy = makeLumps(racks, units)
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
            for (var i=0; i < self.lumpy.length; i++) {
                for (var k=0; k < self.lumpy[i].units.length; k++) {
                    var unit = self.lumpy[i].units[k]
                    if (unit.Mgmt && unit.Mgmt.length > 0) {
                        self.lumpy[i].units[k].pingMgmt = "*"
                    }
                    if (unit.IPs && unit.IPs.length > 0)
                        self.lumpy[i].units[k].pingIP = "*"
                }
            }
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
        "STI": function(val, oldVal) {
                router.push("/rack/layout/" + val)
        },
        "$route": "loadSelf"
    },
})


var userLogin = Vue.component("user-login", {
    template: "#tmpl-user-login",
    data: function() {
        return {
            username: "",
            password: "",
            placeholder: "first.last@pubmatic.com",
            errorMsg: ""
        }
    },
    created: function() {
        console.log("userLogin created")
    },
    methods: {
        cancel: function() {
            console.log("userLogin canceled")
            router.push("/")
        },
        login: function(ev) {
            var data = {Username: this.username, Password: this.password};
            posty(loginURL, data).then(user => {
                this.$store.commit("setUser", user)
                router.push("/")
            }).catch(msg => this.errorMsg = msg.Error)
        },
    },
})


var userLogout = Vue.component("user-logout", {
    template: "#tmpl-user-logout",
    created: function() {
        console.log("userLogout created")
    },
    methods: {
        cancel: function() {
            router.push("/")
        },
        logout: function(ev) {
            console.log("logout button selected")
            this.$store.commit("logOut")
            router.push("/auth/login")
        },
    }
})


// grid component with paging and sorting
var pagedGrid = Vue.component("paged-grid", {
    template: "#tmpl-paged-grid",
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
              sortKey: "",
              sortOrders: sortOrders,
              currentRow: this.startRow
        }
    },
    computed: {
        rowStatus: function() {
            if (! this.rowsPerPage) {
                return this.data.length + ((this.data.length === 1) ? " row" : " rows")
            }
            var status =
                " Page " +
                (this.currentRow / this.rowsPerPage + 1) +
                " / " +
                (Math.ceil(this.data.length / this.rowsPerPage));

            if (this.data.length >  this.rowsPerPage) {
                status += " (" + this.data.length + " rows) ";
            }
            return status
        },
        canDownload: function() {
            return (this.data && (this.data.length > 0) && this.filename && (this.filename.length > 0))
        },
        limitBy: function() {
           var data = (this.rowsPerPage > 0) ? this.data.slice(this.currentRow, this.currentRow + this.rowsPerPage) : this.data;
           var orderBy = (this.sortOrders[this.sortKey] > 0) ? "asc" : "desc";
           return (this.sortKey.length > 0) ? _.orderBy(data, this.sortKey, orderBy) : data
        }
    },
    methods: {
        sortBy: function (column) {
            console.log("sort by:",column)
            this.sortKey = column
            this.sortOrders[column] = this.sortOrders[column] * -1
        },
        movePages: function(amount) {
            var row = this.currentRow + (amount * this.rowsPerPage);
            if (row >= 0 && row < this.data.length) {
                this.currentRow = row;
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

            var element = document.createElement("a");
            var ctype = "application/vnd.ms-excel";
            element.setAttribute("href", "data:" + ctype + ";charset=utf-8," + encodeURIComponent(text));
            element.setAttribute("download", filename);

            element.style.display = "none";
            document.body.appendChild(element);

            element.click();

            document.body.removeChild(element);
        },
    }
});


var sessionList = Vue.component("session-list", {
    template: "#tmpl-session-list",
    mixins: [pagedCommon],
    data: function() {
        return {
            filename: "sessions",
            columns: ["TS", "Login", "Remote", "Event"],
            rows: [],
            searchQuery: "",
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function() {
            getSessions().then(s => this.rows = s)
        },
        linkable: function(key) {
            return (key == "Login")
        },
        linkpath: function(entry, key) {
            return "/user/edit/" + entry["USR"]
        }
    }
})


//
// SITE LIST
//
var siteList = Vue.component("site-list", {
    template: "#tmpl-site-list",
    mixins: [pagedCommon],
    data: function() {
        return {
            sites: [],
            searchQuery: "",
            rows: [],
            columns: [
                "Name",
                "City",
                "Country",
                "Note",
            ],
        }
    },
    created: function () {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            get(sitesURL).then(data => this.rows = data)
        },
        linkable: function(key) {
            return (key == "Name")
        },
        linkpath: function(entry, key) {
            return "/site/edit/" + entry["STI"]
        },
    },
})

var siteEdit = Vue.component("site-edit", {
    template: "#tmpl-site-edit",
    data: function() {
        return {
            Site: {},
        }
    },
    created: function() {
        this.loadSelf()
    },
    methods: {
        loadSelf: function () {
            if (this.$route.params.STI > 0) {
                var url = sitesURL + this.$route.params.STI;
                get(url).then(s => this.Site = 1)
            } else {
                this.Site = {STI: 0}
            }
        },
        showList: function() {
            router.push("/site/list")
        },
        deleteSelf: function() {
            deleteIt(sitesURL + this.Site.STI, this.showList)
        },
        saveSelf: function()  {
            saveMe(sitesURL, this.Site, this.Site.STI, this.showList)
        }
    },
})



var homePage = Vue.component("home-page", {
    template: "#tmpl-home-page",
    data: function() {
        return {
            title: "PubMatic Datacenters",
            rows: [],
            columns: [ "Site", "Servers", "VMs" ],
            testData: "this is a test",
        }
    },
    created: function() {
        if (this.$store.getters.loggedIn) {
            this.loadSelf()
        } else {
            router.push("/auth/login")
        }
    },
    methods: {
        loadSelf: function() {
            get(summaryURL).then(data => this.rows = data)
                .catch(x => console.log("oh fuck:", x))
        },
        dl: function() {
            download("test.txt", this.testData)
        }
    },
})

var tallyMIX = {
    methods: {
        rowid: function(entry) {
            return "sti-" + entry.STI
        },
        linkable: function(key) {
            return (key == "Label")
        },
        linkpath: function(entry, key) {
            if (key == "Label") return "/rack/edit/" + entry["RID"]
        }
    },
}

var tallyho = childTable("tally-table", "#tmpl-base-table", [tallyMIX])


const routes = [
{ path: "/auth/login",      component: userLogin },
{ path: "/auth/logout",     component: userLogout },
{ path: "/admin/sessions",  component: sessionList },
{ path: "/admin/tags",      component: tagEdit },
{ path: "/ip/edit/:IID",    component: ipEdit },
{ path: "/ip/list",         component: ipList },
{ path: "/ip/reserve",      component: ipReserve },
{ path: "/ip/types",        component: ipTypes },
{ path: "/ip/type/edit/:IPT", component:  ipTypeEdit },
{ path: "/ip/reserved",     component:  reservedIPs},
{ path: "/vlan/edit/:VLI",  component:  vlanEdit },
{ path: "/vlan/list",       component: vlanList },
{ path: "/device/add/:STI/:RID/:RU", component: deviceAdd, name: "device-add" },
{ path: "/device/audit/:DID", component: deviceAudit },
{ path: "/device/edit/:DID", component: deviceEdit },
{ path: "/device/list", component:  deviceList },
{ path: "/device/load", component: deviceLoad },
{ path: "/device/types", component:  deviceTypes },
{ path: "/device/type/edit/:DTI", component:  deviceTypeEdit },
{ path: "/vm/audit/:VMI", component: vmAudit },
{ path: "/vm/edit/:VMI", component: vmEdit },
{ path: "/vm/add/:DID", component: vmAdd },
{ path: "/vm/list",         component:  vmList },
{ path: "/mfgr/edit/:MID",  component: mfgrEdit },
{ path: "/mfgr/list",       component:  mfgrList },
{ path: "/part/add",        component:  partEdit },
{ path: "/part/edit/:PID",  component: partEdit },
{ path: "/part/list/:STI",  component: partList },
{ path: "/part/load",           component:  partLoad },
{ path: "/part/type/edit/:PTI", component:  partTypeEdit },
{ path: "/part/type/list",          component:  partTypes },
{ path: "/part/use/:STI/:KID", component: partUse },
{ path: "/part/inventory", component: partInventory },
{ path: "/rack/edit/:RID", component: rackEdit },
{ path: "/rack/list", component: rackList },
{ path: "/rack/layout/:STI", component: rackLayout },
{ path: "/rack/layout", redirect: "/rack/layout/0" },
{ path: "/rma/create/:PID", component:  rmaCreate },
{ path: "/rma/edit/:RMD", component:  rmaEdit },
{ path: "/rma/list", component:  rmaList },
{ path: "/site/edit/:STI", component:  siteEdit },
{ path: "/site/list", component:  siteList },
{ path: "/user/edit/:USR", component:  userEdit },
{ path: "/user/list", component:  userList },
{ path: "/vendor/edit/:VID", component:  vendorEdit },
{ path: "/vendor/list", component:  vendorList },
{ path: "/search/:searchText", component:  searchFor, name: "searcher" },
{ path: "/", component: homePage },
]


// load user info from cookie if it exists
const checkUser = fromCookie();
if (checkUser) {
    checkUser['COOKIE'] = true;
    store.commit("setUser", checkUser)
    //console.log("user is set");
}

const router = new VueRouter({
   routes // short for routes: routes
})

var app = new Vue({
    router,
    store
}).$mount("#myapp")


router.beforeEach((to, from, next) => {
    if (store.getters.loggedIn || to.path == "/auth/login") {
        next()
    } else {
        next("/auth/login")
    }
})

