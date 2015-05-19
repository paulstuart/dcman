var containsAllAscii = function (str) {
    return  /^[\000-\177]*$/.test(str) ;
};


var goback = function () {
    /* no need to stick around - go back to prior page */
    console.log('going back now');
    history.back();
};

var page_refresh = function() {
    window.location = window.location.href; 
};

var update_confirmation = function(data, textStatus, jqXHR) {
    $('#cmd_status').removeClass('cmd_warn');
    if (data.status === "ok") {
        $('#cmd_status').addClass('cmd_ok');
    } else {
        $('#cmd_error').addClass('cmd_error');
    }
    var msg = data.msg + " [" + data.status + "]";
    $('#cmd_status').html(data.msg);
};

function isNumber (o) {
      return ! isNaN (o-0);
}


function fixDate(val) {
    // already date formatted?
    if ((parseInt(val) > 0) && (! isNumber(val))) {
        return val;
    }
    // nada
    if ((parseInt(val) === 0) || val.match(/none/i)) { 
        return "";
    }
    var d = new Date(val * 1000);
    var when = d.toISOString();
    when = when.substring(0,10) + " " + when.substring(11,16);
    return when; 
}

var get_type = function (thing){
    if(thing===null)return "[object Null]"; // special case
    return Object.prototype.toString.call(thing);
}

/**
 * Convert number of bytes into human readable format
 *
 * @param integer bytes     Number of bytes to convert
 * @param integer precision Number of digits after the decimal separator
 * @return string
 */

function bytesToSize(bytes, precision)
{   
    /* default precision of 2 places */
    precision = typeof precision !== 'undefined' ? precision : 2;

    if ((typeof bytes == 'string') && ((parseInt(bytes) === 0) || bytes.match(/none/i))) { 
        return '';
    }
    if (bytes == 'None' || bytes == 'undefined') {
        return '';
    }
    if (! isNumber(bytes)) {
        return bytes;
    }
    var kilo = 1024;
    var mega = kilo * 1024;
    var giga = mega * 1024;
    var tera = giga * 1024;
    
    if (bytes.length == 0) {
        return "";
    } else if ((bytes >= 0) && (bytes < kilo)) {
        return bytes + ' B';
    } else if ((bytes >= kilo) && (bytes < mega)) {
        return (bytes / kilo).toFixed(precision) + ' KB';
    } else if ((bytes >= mega) && (bytes < giga)) {
        return (bytes / mega).toFixed(precision) + ' MB';
    } else if ((bytes >= giga) && (bytes < tera)) {
        return (bytes / giga).toFixed(precision) + ' GB';
    } else if (bytes >= tera) {
        return (bytes / tera).toFixed(precision) + ' TB';
    }
    return bytes;
}

var sizeToBytes = function(size) {
    var r = /(\d+)\s*([kmgt])b?/gi;
    var m = size.split(r);
    if (m.length === 1) { return parseInt(size); }
    var size = parseInt(m[1]);
    switch (m[2].toUpperCase()) {
    case 'K':
        return size * 1024;
        break;
    case 'M':
        return size * 1024 * 1024;
        break;
    case 'G':
        return size * 1024 * 1024 * 1024;
        break;
    case 'T':
        return size * 1024 * 1024 * 1024 * 1024;
        break;
    default:
        return NaN;
    }
};
