 
var get = function(url) {
  // Return a new promise.
  return new Promise(function(resolve, reject) {
    // Do the usual XHR stuff
    var req = new XMLHttpRequest();
    req.open('GET', url);
    if (user_apikey && user_apikey.length > 0) {
	    req.setRequestHeader("X-API-KEY", user_apikey)
    }

    req.onload = function() {
      // This is called even on 404 etc
      // so check the status
      console.log('get status:', req.status, 'txt:', req.statusText)
      if (req.status == 200) {
        // Resolve the promise with the response text
        //console.log('woo:', req.responseText)
        var obj = JSON.parse(req.responseText)
        resolve(obj)
        //resolve(JSON.parse(req.responseText))
        //resolve(req.response);
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

var getData = function(url) {
  return new Promise(function(resolve, reject) {
    // Do the usual XHR stuff
    var req = new XMLHttpRequest();
    req.open('GET', url);
    if (user_apikey && user_apikey.length > 0) {
	    req.setRequestHeader("X-API-KEY", user_apikey)
    }

    req.onload = function() {
      // This is called even on 404 etc
      // so check the status
      //console.log('get status:', req.status, 'txt:', req.statusText)
      if (req.status == 200) {
        var obj = JSON.parse(req.responseText)
        resolve(req)
      }
      else {
        // Otherwise reject with the status text
        // which will hopefully be a meaningful error
        console.log('get failed. url:', url, 'status:',req.status, 'txt:', req.statusText)
        reject(Error(req.statusText));
      }
    };

    // Handle network errors
    req.onerror = function() {
      console.log('get network error for url:',url);
      reject(Error("Network Error"));
    };

    // Make the request
    req.send();
  });
}

var getJSON = function(url) {
      return getData(url).then(JSON.parse);
}

var postIt = function(url, data, fn, method) {
    var xhr = new XMLHttpRequest();
    if (typeof method == "undefined") var method = 'POST';
    xhr.open(method, url, true);
    //xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded; charset=UTF-8');
	xhr.setRequestHeader("Content-Type", "application/json")
    if (user_apikey && user_apikey.length > 0) {
	    xhr.setRequestHeader("X-API-KEY", user_apikey)
    }
    xhr.send(JSON.stringify(data));
    xhr.onreadystatechange = function() {
        if (typeof fn === 'function') {
            fn(xhr);
            return
        }
/*
        if (xhr.readyState == 4 && xhr.status != 200) {
            alert("Oops:" + xhr.responseText);
        }
*/
    };
}

var deleteIt = function(url, fn) {
    postIt(url, null, fn, 'DELETE')
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
/*
    xhr.setRequestHeader("Content-length", form.length);
    xhr.setRequestHeader("Connection", "close");
*/

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

var fetchData = function (url, fn) {
      var xhr = new XMLHttpRequest()
      xhr.open('GET', url)
      //xhr.setRequestHeader("Access-Control-Allow-Origin", "*")
      if (user_apikey && user_apikey.length > 0) {
	      xhr.setRequestHeader("X-API-KEY", user_apikey)
      }
      xhr.onload = function () {
        if (fn) fn(JSON.parse(xhr.responseText), xhr.status)
      }
      xhr.send()
}

// fetch synchronously
var fetchNow = function (url, fn) {
      var xhr = new XMLHttpRequest()
      xhr.open('GET', url, false)
      xhr.onload = function () {
        fn(JSON.parse(xhr.responseText), xhr.status)
      }
      xhr.send()
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
              sortOrders: {} //sortOrders
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

/****** Move this to 'common.js' or once figured out *******/

/*
Vue.component('main-menu', function (resolve, reject) {
    var url = 'static/html/menutmpl.html';
    var xhr = new XMLHttpRequest();
    xhr.open('GET', url)
    xhr.onload = function () {
        if (xhr.status === 200) {
            var parser = new DOMParser();
            var doc = parser.parseFromString(xhr.responseText, "text/html");
            resolve({
                template: doc,
                props: ['app', 'hey', 'msg'],
            });
        }
    }
    xhr.send()
});
*/

// TODO: generate from cookie data
var menuMIX = {
  data: {
      hello: "my name is waldo",
      hey: "what's up, duuuuude?",
      msg: "secret message",
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

var inURL = "/dcman/api/inventory/";
var serverURL = "/dcman/api/server/view/";
var vmURL = "/dcman/api/vm/";
var vmViewURL = "/dcman/api/vm/view/";
var partTypesURL = "/dcman/api/part/type/";
var partURL = "/dcman/api/part/view/";
var rackURL = "/dcman/api/rack/view/";
var rmaURL = "/dcman/api/rma/";
var rmaviewURL = "/dcman/api/rma/view/";
var tagURL = "/dcman/api/tag/";
var sitesURL = "/dcman/api/site/" ; 
var networkURL = "/dcman/api/network/ip/used/";
var userURL = "/dcman/api/user/" ; 
var vlanURL = "/dcman/api/vlan/view/" ; 
var vendorURL = "/dcman/api/vendor/" ; 

var RMA = function() {
    Maker(this, [
        'RMAID',
        'SID',
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
        'UID',
    ])
}

var Part = function() {
    Maker(this, [
        'PID',
        'DID',
        'STI',
        'PTI',
        'Site',
        'Hostname',
        'Description',
        'PartNumber',
        'Serial',
        'AssetTag',
        'Mfgr',
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
        'VIP',
        'Profile',
        'Note',
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
        'ID',
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

var Device = function() {
    Maker(this, [
        'Alias',
        'AssetTag',
        'Assigned',
        'Site',
        'DID',
        'DTI',
        'DevType',
        'Height',
        'Hostname',
        'ID',
        'Note',
        'PartNo',
        'Profile',
        'Rack',
        'RID',
        'RU',
        'SerialNo',
        'STI',
        'TID',
    ])
}
