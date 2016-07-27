
var serverLIST = [];
var siteLIST = [];

var validIP = function(ip) {
    var octs = ip.split('.')
    if (octs.length != 4) return false
    for (var i=0; i<4; i++) {
        var val=parseInt(octs[i])
        if (val < 0 || val > 255) return false
    }
    return true 
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





var foundVue = {
    data: function () {
        return {
            columns: ['Kind', 'Name'],
            rows: [],
            what: '',
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
            if (entry.Kind == 'server') {
                return '/server/edit/' + entry['ID']
            }
            return '/vm/edit/' + entry['ID']
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
            this.rows = funny
        }
    }
}


var fList = Vue.component('found-list', {
    //template: '#tmpl-found-table',
    template: '#tmpl-base-table',
    mixins: [foundVue],
})

Vue.component('my-nav', {
    template: '#tmpl-main-menu',
    props: ['app', 'msg'],
    data: function() {
       return {
           searchText: 'platform9',
           found: [],
           columns: ['Kind', 'Name'],
       }
    },
    created: function() {
        this.userinfo()
    },
    methods: {
        'doSearch': function(ev) {
            console.log('search for:',this.searchText)
            this.$dispatch('search-for', this.searchText)
        },
        'userinfo': function() {
            var cookies = document.cookie.split("; ");
            for (var i=0; i < cookies.length; i++) {
                var tuple = cookies[i].split('=')
                if (tuple[0] != 'userinfo') continue;
                var user = JSON.parse(atob(tuple[1]));
                console.log("***** PRE:", tuple[1])
                console.log("***** USER:", user)
                this.$dispatch('user-info', user)
                break
            }
        },
    'XXdoSearch': function(ev) {
        var self = this;
        //console.log("SEARCH:",this.searchText)
        if (this.searchText.length > 0) { 
            var searchURL = "/dcman/api/search/";
            var url = searchURL + this.searchText;
            fetchData(url, function(data) {
                //console.log('search completed')
                if (data) {
                    console.log('search for:', self.searchText, 'm-atched:', data.length)
                    if (data.length == 1) {
                        if (data[0].Kind == 'server') {
                            router.go('/server/edit/' + data[0].ID)
                            self.$dispatch('server-found', 'oh pretty please')
                            return
                        }
                        if (data[0].Kind == 'vm') {
                            router.go('/vm/edit/' + data[0].ID)
                            self.$dispatch('vm-found', 'yes please')
                            return
                        }
                        alert("unknown kind:",data[0].Kind)
                    } else {
                        //self.found = data
                        /*
                        router.go({
                            name: 'found',
                            //params: {got: this.searchText}
                            params: 'bad luck woman'
                        })
                        */
                        router.go('/found')
                            /*
                        self.$dispatch('found-these', data)
                        self.$emit('found-these', data)
                        */
                        self.$broadcast('found-these', data)
                    }
                }
            })
        }
    }
  }
})


var ipload = {
    methods: {
        loadData: function() {
            console.log("let all us load our data!")
            var self = this,
                 url = networkURL;
            if (self.DCD > 0) {
                url +=  "?DCD=" + self.DCD;
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
    props: [ 'DCD', 'Kind', 'What'],
    created: function(ev) {
        console.log('t estmix created!');
    },
    ready: function(ev) {
        console.log('t estmix ready!');
    },
    methods: {
        subFilter: function(a, b, c) {
            if (! this.What && ! this.Kind) {
                return a
            }
            if (this.What == a.What && ! this.Kind) {
                return a
            }
            if (this.Kind == a.Kind && ! this.What) {
                return a
            }
            if (this.Kind == a.Kind && this.What == a.What) {
                return a
            }
        },
        linkable: function(key) {
            return (key == 'Hostname')
        },
        linkpath: function(entry, key) {
            if (entry.Kind == 'server') {
                return '/server/edit/' + entry['ID']
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
                "DC",
                "Kind",
                "What",
                "Hostname",
                "IP",
                "Note"
            ],
            sortKey: '',
            sortOrders: [],
            What: '',
            Kind: '',
            DCD: 1,
            searchQuery: '',
            dcs: [],
            whatlist: [
               '',
                'ipmi',
                'internal',
                'public',
                'vip',
            ],
            kindlist: [
               '',
                'vm',
                'server',
            ],
        }
    },
    attached: function () {
        this.dcs = siteLIST;
    },
    ready: function(ev) {
        console.log('t estmix ready!');
    },
    events: {
        'ip-reload': function(msg) {
            console.log("reload those IPs!!!!!: ", msg)
        }
    },
    watch: {
        'DCD': function(x) {
            this.loadData()
        }
    },
}


var dg = childTable("ipgrid", "#tmpl-base-table", [ipgridMIX])

var Netlist = Vue.component('netlist', {
    template: '#ip-table',
    mixins: [ipload, ippageMIX],
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
/*
var userEditVue = {
    data: function() {
        return {
            User: new(User),
            url: userURL,
        }
    },
    created: function () {
        this.loadSelf()
    },

    methods: {
        myID: function() {
            return this.User.ID
        },
        myself: function() {
            return this.User
        },
        saveSelf: function() {
            var data = this.myself()
            var id = this.myID()
            if (id > 0) {
                postIt(this.url + id + "?debug=true", data, this.showList, 'PATCH')
            } else {
                postIt(this.url + id + "?debug=true", data, this.showList)
            }
        },
        deleteSelf: function(event) {
            console.log('delete event: ' + event)
            postIt(this.url + this.myID(), null, this.showList, 'DELETE')
        },
        showList: function(ev) {
            //ev.preventDefault();
            router.go('/user/list')
        },
        loadSelf: function () {
            var self = this;
            var id = this.$route.params.UID;
            if (id > 0) {
                var url = userURL + id;

                fetchData(url, function(data) {
                    self.User.Load(data);
                })
            }
        },
    },
}
*/
var userEditVue = {
    data: function() {
        return {
            User: new(User),
            dataURL: userURL,
            listURL: '/user/list',
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
// Server Edit
//
var serverEditVue = {
  data: function() {
      return {
        dcs: [],
        racks: [],
        tags: [],
        SID: 0,
        Description: '',
        Server: new(Server),
    }
  },
  created: function () {
      this.loadTags()
      this.loadServer()
  },
    attached: function() {
        console.log("DEVICE ATTACHED:", this.$route.params.SID)
         var id = this.$route.params.SID;
        if (id && id != this.SID) {
            this.loadServer()
        }

        this.dcs = siteLIST;
    },
    methods: {
        Updated: function(event) {
            console.log('send update event: ' + event);
            this.Server.TID = parseInt(this.Server.TID)
            postIt(serverURL + this.SID + "?debug=true", this.Server, this.showList, 'PATCH')
        },
        Deleted: function(event) {
            console.log('delete event: ' + event)
            postIt(serverURL + this.SID, null, this.showList, 'DELETE')
        },
        showList: function(ev) {
            router.go('/server/list')
        },
        loadRacks: function () {
             var self = this;
             fetchData(rackURL + "?DCD=" + self.Server.DCD, function(data) {
                 self.racks = data
             })
        },
        loadServer: function () {
             var self = this;

             var id = this.$route.params.SID;
             console.log('loading server ID:', id)
             if (id > 0) {
                 this.SID = id
                 var url = serverURL + id;

                 fetchData(url, function(data) {
                     self.Server.Load(data);
                     self.loadRacks()
                 })
               }
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
        "Server.DCD": function(older, newer) {
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
/*
        myID: function() {
            return this.User.ID
        },
        myself: function() {
            return this.User
        },
*/

var serverEdit = Vue.component('server-edit', {
    template: '#server-edit-template',
    mixins: [serverEditVue],
})


//
// VM Edit
//
var vmEditVue = {
  data: function() {
      return {
          url: vmURL,
        DCD: 0,
        dcs: [],
        racks: [],
        tags: [],
        Description: '',
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
        this.dcs = siteLIST;
    },

    methods: {
        saveSelf: function(event) {
            console.log('send update event: ' + event);
            //this.Server.TID = parseInt(this.Server.TID)
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
/*            
             fetchData(dcURL, function(data) {
                 self.dcs = data
             })
*/
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
            //console.log('* reload server ID:', id, 'current ID:', thisVM.VMI, 'VMI:', this.VMI)
            if (id != this.VM.VMI) {
                console.log('********************** NEW server ID:', id)
                this.loadVM()
            }
        }
    }
}

var vmEdit = Vue.component('vm-edit', {
    template: '#tmpl-vm-edit',
    mixins: [vmEditVue],
})

// Base APP component, this is the root of the app
var App = Vue.extend({
    data: function(){
        return {
            found: [],
            columns: ['Kind', 'Name'],
            what: '',
            multiples: false,
            myapp: {
                auth: {
                    loggedIn: true,
                    user: {
                        name: "Waldo"
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
        search: function(what) {
            var self = this;
            console.log("DO SEARCH:",what)
            this.what = what
            if (what.length > 0) { 
                var searchURL = "/dcman/api/search/";
                var url = searchURL + what;
                fetchData(url, function(data) {
                    //console.log('search completed')
                    if (data) {
                        console.log('search for:', what, 'matched:', data.length)
                        if (data.length == 1) {
                            self.multiples = false
                            console.log('what:', what)
                            if (data[0].Kind == 'server') {
                                router.go('/server/edit/' + data[0].ID)
                                //self.$dispatch('server-found', 'oh pretty please')
                                return
                            }
                            if (data[0].Kind == 'vm') {
                                router.go('/vm/edit/' + data[0].ID)
                                self.$dispatch('vm-found', 'yes please')
                                return
                            }
                            alert("unknown kind:",data[0].Kind)
                        } else {
                            router.go('/found')
                            self.$broadcast('found-these', data)
                            console.log('broadcasting to found-these')
                        }
                    }
                })
            }
        }
    },
    events: {
        'server-found': function (ev) {
            console.log('app reload event:', ev)
            this.$broadcast('server-reload', 'gotcha!')
        },
        'user-info': function (user) {
            console.log('*** user-info event:', user)
            this.myapp.auth.user.name = user.username;
        },
        'search-for': function (msg) {
            console.log('*** search-for event:', msg)
            this.search(msg)
        },
        'hide-found': function(msg) {
            console.log('*** hide-found event:', msg)
            this.multiples = false
        }
    },
})

var serverMIX = {
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
            return '/server/edit/' + entry['ID']
        }
    },
}

var sg = childTable("server-grid", "#tmpl-base-table", [serverMIX])

var serverListVue = {
  data: function() {
      return {
      DCD: 1,
      RID: 0,
      dcs: [],
      racks: [],
      searchQuery: '',
      gridData: [],
      gridColumns: [
           "DC",
           "Rack",
           "RU",
           "Hostname",
           "IPInternal",
           "IPIpmi",
           "Tag",
           "Profile",
           "Serial",
           "AssetTag",
           "Assigned",
           "Note",
        ]
    }
  }
  ,
  created: function () {
      this.loadSelf()
  },
  attached: function () {
      this.dcs = siteLIST;
  },
  methods: {
    loadRacks: function () {
         var self = this,
              url = rackURL + "?DCD=" + self.DCD;

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


         var url = serverURL;
         var dcd = parseInt(self.DCD);
         if (dcd > 0) {
             url += "?dcd=" + self.DCD
         }
         fetchData(url, function(data) {
             self.gridData = data
         })
         self.loadRacks()
    },
  },

  watch: {
    'DCD': function(val, oldVal){
            this.loadSelf()
        },
    },
}

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
            dcs: [],
            searchQuery: '',
            rows: [],
            columns: [
                "DC",
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
    attached: function () {
        this.dcs = siteLIST;
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
    mixins: [vlanListVue],
})

//
// VLAN Edit
//
var vlanEditVue = {
    data: function() {
        return {
            dcs: [],
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

          this.dcs = siteLIST;
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
    mixins: [editVue, vlanEditVue],
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
            if (key == 'Server') return '/server/edit/' + entry['SID']
            if (key == 'Hostname') return '/vm/edit/' + entry['VMI']
        }
    },
}
var vg = childTable("vm-grid", "#tmpl-base-table", [vmMIX])

var vmListVue = {
  data: function() {
      return {
      DCD: 1,
      dcs: [],
      searchQuery: '',
      gridData: [],
      gridColumns: [
           "DC",
           "Server",
           "Hostname",
           "Private",
           "Public",
           "VIP",
           "Profile",
           "Note",
        ]
    }
  }
  ,

  created: function () {
      this.loadSelf()
  },
    attached: function () {
        this.dcs = siteLIST;
    },

  methods: {
    loadRacks: function () {
         var self = this,
              url = rackURL + "?DCD=" + self.DCD;

         fetchData(url, function(data) {
             if (data) {
                 data.unshift({RID:0, Label:''})
                 self.racks = data
             }
         })
    },
    loadSelf: function () {
         var self = this
        // self.RID = 0

         var url = vmURL;
         var dcd = parseInt(self.DCD);
         if (dcd > 0) {
             url += "?dcd=" + self.DCD
         }
         fetchData(url, function(data) {
             self.gridData = data
         })
         self.loadRacks()
    },
  },

  watch: {
    'DCD': function(val, oldVal){
            this.loadSelf()
        },
    },
}

var ServerList = Vue.component('server-list', {
    template: '#tmpl-server-page',
    mixins: [serverListVue],
})

var VMList = Vue.component('vm-list', {
    template: '#tmpl-vm-page',
    mixins: [vmListVue],
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
            return '/part/inventory/' + entry['DCD'] + '/' + entry['KID']
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

inventoryVue = {
  data: function() {
      return {
      showgrid: true,
      SID: 0,
      DCD: 4,
      PID: 0,
      available: [],
      dcs: [],
      hostname: '',
      other: '',
      searchQuery: '',
      partData: [],
      gridColumns: ['DC', 'Description', 'Mfgr', 'Qty'],
    }
  },
  created: function () {
      this.loadSelf()
  },
    attached: function () {
        this.dcs = siteLIST;
    },

  methods: {
    updated: function(event) {
        console.log('the event: ' + event)
    },
    loadSelf: function () {
         var self = this;
         var url = inURL;
         if (self.DCD > 0) {
            url += "?dcd=" + self.DCD
         }
         fetchData(url, function(data) {
             self.partData = data
         })
     },
  },

  watch: {
    'DCD': function(val, oldVal){
            this.loadSelf()
        },
    },
}


var iiPart = Vue.component('part-inventory', {
    template: '#tmpl-inventory-list',
    mixins: [inventoryVue],
})


partUseVue = {
  data: function() {
      return {
      SID: 0,
      DCD: 4,
      PID: 0,
      available: [],
      dcs: [],
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
          var dcd = this.$route.params.DCD;
          var url = partURL + "?unused=1&bad=0&kid=" + kid + "&dcd=" + dcd
          fetchData(url, function(data) {
              self.available = data
          })
      },
      thisPart: function(ev) {
          //alert("ok!")
          var pid = document.getElementById("PID").value
          var part = {
            PID: parseInt(pid),
            DCD: this.DCD,
            SID: this.SID,
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
                self.SID = enable ? data.ID : 0;
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

partListVue = {
  data: function() {
      return {
      showgrid: true,
      isgood: true,
      isbad: false,
      SID: 0,
      DCD: 4,
      PID: 0,
      available: [],
      dcs: [],
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
           "DC",
           "Hostname",
           "Serial",
           "PartType",
           "PartNumber",
           "Description",
           "Mfgr",
           "Bad",
        ]
      }
    }
  ,
  created: function () {
      this.loadSelf()
  },
  attached: function () {
      this.dcs = siteLIST;
  },
  methods: {
    loadSelf: function () {
         var self = this
         var url = partURL;
        if (self.DCD > 0) {
         //if (parseInt(self.DCD) > 0) {
            url += "?dcd=" + self.DCD
         }
         //var url = partURL + "?dcd=" + self.DCD
         fetchData(url, function(data) {
            if (data) {
            for (i=0; i<data.length; i++) {
                data[i].Serial = data[i].Serial || '';
            }
            }
             self.partData = data
         })
    },
      partOk: function(ev) {
          //alert("ok!")
          var part = {
            DCD: parseInt(this.DCD),
            KID: parseInt(document.getElementById("new-kid").value),
            Serial: document.getElementById("add-sn").value,
            Unused: 1,
          }
          postIt(mypartURL, part)
      },
      useOk: function(ev) {
          //alert("ok!")
          var pid = document.getElementById("PID").value
          var part = {
            PID: parseInt(pid),
            DCD: this.DCD,
            SID: this.SID,
            Unused: 0,
          }
          postIt(mypartURL + pid, part, null, "PATCH")
      },
      partUsed: function(ev) {
          //alert("nope!")
      },
      partNope: function(ev) {
          //alert("nope!")
      },
      findhost: function(ev) {
          var self = this;
          console.log("find hostname:",this.hostname);
            fetchData("api/server/hostname/" + this.hostname, function(data, status) {
                var enable = (status == 200);
                buttonEnable(document.getElementById('use-btn'), enable)
                self.SID = enable ? data.ID : 0;
            })
      },
      newPart: function(ev) {
        var id = parseInt(ev.target.id.split('-')[1]);
      },
  },


  events: {
      'new-item': function(id) {
          // would be nice to have hash lookup
            for (i = 0; i < this.partData.length; i++) {
                if (id == this.partData[i].KID) {
                    document.getElementById("add-desc").value = this.partData[i].Description
                    document.getElementById("add-mfgr").value = this.partData[i].Mfgr
                    document.getElementById("new-kid").value = id
                    break
                }
            }
            document.getElementById("table-only").style.display = "none"
            document.getElementById("added-part").style.display = "inline"
            //console.log('NEW ITEM:' + data)
      },
      'item-added': function (msg) {
        this.loadSelf()
      },
      'select-part': function (id) {
            console.log("use part id:",id);
            var self = this;
            var url = mypartURL + "?unused=1&bad=0&kid=" + id
            postIt(url, null, function(xreq) {
                if (xreq.readyState == 4) {
                     if (xreq.status != 200) {
                        alert("Oops:" + xreq.responseText);
                        return
                     }
                     //console.log(xreq.responseText)
                     self.available = JSON.parse(xreq.responseText)

                    document.getElementById("table-only").style.display = "none"
                    document.getElementById("use-part").style.display = "inline"
                }
            }, 'GET')
      },
   },

}


var iList = Vue.component('parts-inventory', {
    template: '#tmpl-inventory-list',
    mixins: [inventoryVue],
})



var partsMIX = {
    methods: {
      noGood: function(a, b, c) {
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
Vue.component('parts-table', {
  //template: '#tmpl-parts-table',
  template: '#tmpl-base-table',
  props: ['kfilter', 'filterKey', 'myfn'],
  mixins: [tableTmpl, partsMIX],
})

//
// All Parts
//
var pList = Vue.component('part-list', {
    template: '#tmpl-part-list',
    mixins: [partListVue],
})

//
// USE PART
//

//
// RMAs
//

// register the grid component

var rmaListMIX = {
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
        }
    },
}


var rmaListVue = {
  data: function() {
      return {
      DCD: 4,
      dcs: [],
      rmas: [],
      searchQuery: '',
      gridColumns: [
        "Description",
        "Hostname",
        "ServerSN",
        "PartSN",
        "VendorRMA",
        "Jira",
        "Shipped",
        "Received",
        "Closed",
        "Created",
        ]
    }
  },
  created: function () {
      this.loadSelf()
  },
  attached: function () {
      this.dcs = siteLIST;
  },
  methods: {
    loadSelf: function () {
         var self = this;
         fetchData(rmaviewURL + "?DCD=" + self.DCD, function(data) {
             self.rmas = data
         })
    },
  },
  watch: {
    'DCD': function(val, oldVal){
            this.loadSelf()
        },
    },
}

var rg = childTable("rma-grid", "#tmpl-base-table", [rmaListMIX])

var rList = Vue.component('rma-list', {
    template: '#tmpl-rma-list',
    mixins: [rmaListVue],
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
// PART EDIT
//
var partEditVue = {
    data: function() {
       return {
            dcs: [],
            DCD: 4,
            Part: new(Part),
       }
    },
    created: function () {
        this.loadPart()
    },
    attached: function () {
        this.dcs = siteLIST;
    },
    route: {
        activate: function () {
            console.log('====== part route activated! =========')
        },
    },
    methods: {
        showList: function(ev) {
            router.go('/part/list')
        },
        saveSelf: function(event) {
            console.log('send event: ' + event)
            var url = partURL + this.Part.PID;
            url += "?debug=true"
            postIt(url, this.Part, this.showList, 'PATCH')
        },
        addSelf: function(event) {
            console.log('add event: ' + event)
            var url = partURL + this.Part.PID;
            url += "?debug=true"
            postIt(url, this.Part, this.showList)
        },
        doRMA: function(ev) {
            ev.preventDefault();
            var data = {
                DCD: this.DCD,
                OldPID: this.Part.PID,
            };
            postIt(rmaURL + "?debug=true", data, function(xhr) {
                if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 201)) {
                    var rma = JSON.parse(xhr.responseText)
                    //console.log("XREF: ", xhr.responseText)
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
                    self.Part.SID = enable ? data.ID : 0;
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
        dcs: [],
        racks: [],
        tags: [],
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
      showList: function() {
          //ev.preventDefault();
          router.go('/rma/list')
      },
      loadRMA: function () {
          var self = this;
          var id = this.$route.params.RMAID;
          if (id > 0) {
              this.SID = id
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
                self.RMA.SID = enable ? data.ID : 0;
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
        return {
            tags: [],
            url: tagURL,
            tag: new(Tag),
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
    },
    watch: {
        'tag.TID': function() {
            for (var i=0; i < this.tags.length; i++) {
                if (this.tags[i].TID == this.tag.TID) {
                    this.tag.Name = this.tags[i].Name
                    return
                }
                this.tag.Name = ''
            }
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
            dcs: [],
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
    attached: function() {
        console.log("RACK ATTACHED:", this.$route.params[this.id])
        var id = this.$route.params[this.id];
        if (id && id != this.myID()) {
              this.loadSelf()
        }
        this.dcs = siteLIST;
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
    mixins: [editVue, rackEditVue],
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
var vg = childTable("rack-grid", "#tmpl-base-table", [rackMIX])

var rackListVue = {
    data: function() {
        return {
            dataURL: '/dcman/api/rack/view/',
            DCD: 1,
            dcs: [],
            searchQuery: '',
            rows: [],
            columns: [
               "DC",
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
    attached: function () {
        this.dcs = siteLIST;
    },

    methods: {
        loadSelf: function () {
             var self = this

             var url = this.dataURL;
             var dcd = parseInt(self.DCD);
             if (dcd > 0) {
                 url += "?dcd=" + self.DCD
             }
             fetchData(url, function(data) {
                 self.rows = data
             })
        },
  },

  watch: {
    'DCD': function(val, oldVal){
            this.loadSelf()
        },
    },
}

var rackList = Vue.component('rack-list', {
    template: '#tmpl-rack-list',
    mixins: [rackListVue],
})

var rackLayoutVue = {
    data: function() {
        return {
            dataURL: '/dcman/api/rackunit/',
            DCD: 1,
            RID: 0,
            dcs: [],
            racks: [],
            layouts: [],
        }
    },

    created: function () {
        this.loadSelf()
    },
    attached: function () {
        this.dcs = siteLIST;
    },

    methods: {
        loadSelf: function () {
             var self = this

             var url = this.dataURL;
             var dcd = parseInt(self.DCD);
             var rid = parseInt(self.RID);

             if (rid > 0) {
                 url += "?rid=" + self.RID
             } else if (dcd > 0) {
                 url += "?dcd=" + self.DCD
             }

             fetchData(url, function(data) {
                 self.layouts = data
             })
        },
        ping: function() {
            /*
            var ips = get_ips();
            var ipmis = get_ipmis();

            var iplist = [];
            for (var ip in ips) {
               iplist.push(ip);
            }
            for (var ip in ipmis) {
               iplist.push(ip);
            }
            */
            var url = "http://10.100.182.16:8080/dcman/api/pings?debug=true";
            var ips = [];
/*
            var addIP = function(ip){
               if validIP(ip) ips.push(ip)
            } 
*/
                /*
                if (x.IPMI.length > 0) ips.push(x.IPMI);
                if (x.Internal.length > 0) ips.push(x.Internal);
                */
            for (var i=0; i < this.layouts.length; i++) {
                var x = this.layouts[i];
                if (validIP(x.IPMI)) ips.push(x.IPMI);
                if (validIP(x.Internal)) ips.push(x.Internal);
                // TODO: validate IP is valid
            }
            //postIt(url, {ips: ips}, function(data) {
            var list = ips.join(",");

            postForm(url, {iplist: list}, function(xhr) {
               if (xhr.readyState == 4 && xhr.status == 200) {
                   var pinged = JSON.parse(xhr.responseText)
/*
                    for (var i=0; i < pinged.length; i++) {
                        var what = pinged[i];
                        console.log("PINGED:", what)
                    }
*/
                    for (ip in pinged) {
                        //var mark = pinged[ip] ? "ok" : "-";
                        var cell = document.getElementById("ip-" + ip);
                        cell.innerHTML = (pinged[ip] ? "ok" : "-")
                        
                        //console.log("PINGED:", ip, "OK:", ok)
                        
                    }
               }

                /*
                console.log("PING DATA:", data.length)
                   for (var k in data) {
                        if (data.hasOwnProperty(k)) {
                            var ok = data[k] ? "ok" : "";
                            var ru = ips[k];
                            if (typeof ru !== 'undefined') {
                                $("#ip_ping_"+ru).text(ok);
                                continue;
                            }
                            ru = ipmis[k];
                            if (typeof ru !== 'undefined') {
                                $("#ipmi_ping_"+ru).text(ok);
                            }
                        }
                    }
                */
             });
        }
  },

  watch: {
    'DCD': function(val, oldVal){
            this.loadSelf()
        },
    },
}

// LAYOUT
var rackView = Vue.component('rack-view', {
    template: '#tmpl-rack-view',
    props: ['rack', 'layout', 'layouts'],
})

var rackLayout = Vue.component('rack-layout', {
    template: '#tmpl-rack-layout',
    mixins: [rackLayoutVue],
})

// Assign the new router
var router = new VueRouter()

// Assign your routes
router.map({
    '/admin/tags': {
        component: Vue.component('tag-edit')
    },
    '/ip/list': {
        component:  Vue.component('netlist')
    },
    '/vlan/edit/:VLID': {
        component:  Vue.component('vlan-edit')
    },
    '/vlan/list': {
        component:  Vue.component('vlan-list')
    },
    '/server/edit/:SID': {
        component: Vue.component('server-edit')
    },
    '/server/list': {
        component:  Vue.component('server-list')
    },
    '/vm/edit/:VMI': {
        component: Vue.component('vm-edit')
    },
    '/vm/list': {
        component:  Vue.component('vm-list')
    },
    '/part/edit/:PID': {
        component:  Vue.component('part-edit')
    },
    '/part/list': {
        component:  Vue.component('part-list')
    },
    '/part/inventory/:DCD/:KID': {
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
    '/found': {
        component:  Vue.component('found-list'),
        //name: 'found'
    },
})

fetchData(dcURL, function(data) {
    data.unshift({DCD:0, Name:''})
    siteLIST = data;

    // Start the app
    router.start(App, '#app')
})
