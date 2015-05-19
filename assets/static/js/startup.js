
var getSelectedText = function () {
    if (window.getSelection) {
        return window.getSelection().toString();
    } else if (document.selection) {
        return document.selection.createRange().text;
    }
    return '';
}

var textSelect = function() {
    var text=getSelectedText();
    if (text!='') {
        $("#captured").text(text);
    }
}

var update_page = function( msg, textStatus, jqXHR) {
        var id = "#" + jqXHR.getResponseHeader('X-Line-ID') +  "> .comment_block";
        $(id).append(msg);
  };

var submit_comment = function() {
    var the_comment = {
        org_id: org_id,
        spec_id: spec_id,
        line_id: $('input[name=spec_id]').val(),
        comment: $('#spec_comment').val()
    };
    $.ajax({
      type: "POST",
      url: "/line/comment",
      data: the_comment,
      dataType: "html",
      success: update_page
    });
    toggle_comment();
}

var delete_line = function(id) {
    var data = {
        org_id: org_id,
        spec_id: spec_id,
        line_id: id
    }
    $.ajax({
      type: "POST",
      url: "/line/delete",
      data: data,
      complete: function(){ 
       document.location.reload(); 
     }
    });
}

var undo_action = function() {
    var data = {
        org_id: org_id,
        spec_id: spec_id,
    }
    $.ajax({
      type: "POST",
      url: "/undo",
      data: data,
      complete: function(){ 
       document.location.reload(); 
     }
    });
}

var edit_line = function(id) {
    var text_id = "#" + id + " > .spectext"
    var prior_content =  $(text_id).text().trim()
    $(text_id).attr("contenteditable", "true");
    $(text_id).focus();
    $(text_id).blur(function(ev) {
        //var new_content = $(text_id).text().trim();
        var new_html = $(text_id).html();
        var new_content = new_html.replace(new RegExp("<br>", "g"), '\n').trim();
        /*
        new_text = new_text.replace(new RegExp("<div>(.*)</div>", "g"), '\n' + "$1");
        new_text = new_text.replace(new RegExp("</?div>", "g"), '');
        */
        if (new_content.length == 0) {
            delete_line(id);
        } else if (new_content != prior_content) {
            var data = {
                org_id: org_id,
                spec_id: spec_id,
                line_id: id, 
                text: new_content
            }
            $.ajax({
              type: "POST",
              url: "/line/edit",
              data: data,
              complete: function(){ 
               document.location.reload(); 
             }
            });
        }
        $(text_id).attr("contenteditable", "false");
        ev.stopPropagation();
        /*
        if (this.textContent.trim().length == 0) {
            $("#new_line").remove();
        } else {
        }
        */
    });
}

var demote_line = function(id) {
    var data = {
        org_id: org_id,
        spec_id: spec_id,
        line_id: id 
    }
    $.ajax({
      type: "POST",
      url: "/line/demote",
      data: data,
      complete: function(){ 
       document.location.reload(); 
     }
    });
        /*
        ev.stopPropagation();
        */
}

var promote_line = function(id) {
    var data = {
        org_id: org_id,
        spec_id: spec_id,
        line_id: id 
    }
    $.ajax({
      type: "POST",
      url: "/line/promote",
      data: data,
      complete: function(){ 
       document.location.reload(); 
     }
    });
        /*
        ev.stopPropagation();
        */
}

/*
http://stackoverflow.com/questions/15237304/jquery-content-editable-indent
$(".code").keydown(function(e)
           {
              e = e || window.event;
              var keyCode = e.keyCode || e.which; 
              if (keyCode == 9)
              {
                e.preventDefault();
                  document.execCommand('styleWithCSS',true,null);
                  document.execCommand('indent',true,null);
              }
           });
*/

var last_key = 0;
var old_content = "";

var edit_mode = function() {
/*
    $(".spectext").keypress(function( event ) {
        console.log("pressed:" + event.which);
    });
    $(".spectext").live('keypress', function(e) {
*/
    $(".spectext").attr("contenteditable", "true");
    $(".spectext").focus(function(e) {
        var new_html = e.target.innerHTML;
        old_content = new_html.replace(new RegExp("<br>", "g"), '\n').trim();
    });
    $(".spectext").blur(function(e) {
        var id = e.target.parentNode.id;
        var new_html = e.target.innerHTML;
        var new_content = new_html.replace(new RegExp("<br>", "g"), '\n').trim();
        if (new_content != old_content) {
            /*
            console.log("OLD:" + old_content);
            console.log("NEW:" + new_content);
            */
            if (new_content.length == 0) {
                delete_line(id);
            } else {
                var data = {
                    org_id: org_id,
                    spec_id: spec_id,
                    line_id: id, 
                    text: new_content
                }
                $.ajax({
                  type: "POST",
                  url: "/line/edit",
                  data: data,
                  complete: function(){ 
                   document.location.reload(); 
                 }
                });
            }
            e.stopPropagation();
        }
        /*
        $(text_id).attr("contenteditable", "false");
        */
    });
    $(".spectext").keydown(function(e) {
        console.log("pressed:" + e.which);
        switch(e.which) {
        case 8:
            if (last_key === 16) {
                delete_line(e.target.parentNode.id);
            }
        case 9: 
            var id = e.target.parentNode.id;
            if (last_key === 16) {
                promote_line(id);
            } else {
                demote_line(id);
            }
        case 13:
            append_line(e.target.parentNode.id);
        }
        last_key = e.which;
    });
    /*
       event.preventDefault();
     */
            
}

var insert_line = function(id) {
    id = "#" + id
    $(id).before('<div class="specline" id="new_line"><div class="spectext" contenteditable="true"/></div>');
    $('#new_line > .spectext').focus();
    $("#new_line > .spectext").blur(function(ev) {
        if (this.textContent.trim().length == 0) {
            $("#new_line").remove();
        } else {
            var data = {
                org_id: org_id,
                spec_id: spec_id,
                line_id: id, 
                text: this.textContent
            }
            $.ajax({
              type: "POST",
              url: "/line/insert",
              data: data,
              complete: function(){ 
               document.location.reload(); 
             }
            });
        }
    });
}

var append_line = function(id) {
    id = "#" + id
    $(id).after('<div class="specline" id="new_line"><div class="spectext" contenteditable="true"/></div>');
    $('#new_line > .spectext').focus();
    $("#new_line > .spectext").blur(function(ev) {
        if (this.textContent.trim().length == 0) {
            $("#new_line").remove();
        } else {
            var data = {
                org_id: org_id,
                spec_id: spec_id,
                line_id: id, 
                text: this.textContent
            }
            $.ajax({
              type: "POST",
              url: "/line/append",
              data: data,
              complete: function(){ 
               document.location.reload(); 
             }
            });
        }
    });
}

var under_line = function(id) {
    id = "#" + id
    $(id).after('<div class="specline" id="new_line"><div class="spectext" contenteditable="true"/></div>');
    $('#new_line > .spectext').focus();
    $("#new_line > .spectext").blur(function(ev) {
        if (this.textContent.trim().length == 0) {
            $("#new_line").remove();
        } else {
            var data = {
                org_id: org_id,
                spec_id: spec_id,
                line_id: id, 
                text: this.textContent
            }
            $.ajax({
              type: "POST",
              url: "/line/under",
              data: data,
              complete: function(){ 
               document.location.reload(); 
             }
            });
        }
    });
}

var toggle_comment = function () {
        el = document.getElementById("comment_dialog");
            el.style.visibility = (el.style.visibility == "visible") ? "hidden" : "visible";
}

var make_menu = function() {
    $.contextMenu({
        selector: '.oid', 
        trigger: 'hover',
        autoHide: true,
        delay: 500,
        callback: function(key, options) {
            if (key == "comment") {
                toggle_comment();
                var spec_id = this.context.parentNode.id;
                $('input[name=spec_id]').val(spec_id);
                $('#spec_text').text(this.context.nextSibling.textContent);
            } else {
            var m = "clicked: " + key;
            window.console && console.log(m) || alert(m); 
            }
        },
        items: {
            "comment": {name: "Comment", icon: "comment"},
            "edit": {name: "Edit", icon: "edit"},
            "cut": {name: "Cut", icon: "cut"},
            "copy": {name: "Copy", icon: "copy"},
            "paste": {name: "Paste", icon: "paste"},
            "delete": {name: "Delete", icon: "delete"},
            "sep1": "---------",
            "quit": {name: "Quit", icon: "quit"}
        }
    });
}

var comment_promotion = function(ev) {
    var comment_id = this.parentElement.parentElement.id;
    var the_comment = {
        org_id: org_id,
        spec_id: spec_id,
        comment_id: comment_id
    };
    $.ajax({
      type: "POST",
      url: "/comment/promote",
      data: the_comment,
      complete: function(){ 
       document.location.reload(); 
      }
    });
}

var popupmenu = function() {
    $.contextMenu({
        selector: '.oid', 
        trigger: 'hover',
        autoHide: true,
        delay: 500,
        callback: function(key, options) {
            if (key == "comment") {
                toggle_comment();
                var spec_id = this.context.parentNode.id;
                $('input[name=spec_id]').val(spec_id);
                $('#spec_text').text(this.context.nextSibling.textContent);
            } else if (key == "delete") {
                delete_line(this.context.parentNode.id);
            } else if (key == "insert") {
                insert_line(this.context.parentNode.id);
            } else if (key == "append") {
                append_line(this.context.parentNode.id);
            } else if (key == "under") {
                under_line(this.context.parentNode.id);
            } else if (key == "edit") {
                edit_line(this.context.parentNode.id);
            } else if (key == "demote") {
                demote_line(this.context.parentNode.id);
            } else if (key == "promote") {
                promote_line(this.context.parentNode.id);
            } else {
                var m = "clicked: " + key;
                window.console && console.log(m) || alert(m); 
            }
        },
        items: {
            "comment": {name: "Comment", icon: "comment"},
            "edit": {name: "Edit", icon: "edit"},
            "insert": {name: "Insert", icon: "edit"},
            "append": {name: "Append", icon: "edit"},
            "under": {name: "Insert under", icon: "edit"},
            "promote": {name: "Promote", icon: "edit"},
            "demote": {name: "Demote", icon: "edit"},
            "delete": {name: "Delete", icon: "delete"},
        }
    });
}

$(document).ready(function() {

    $( "#edit_button" ).click(edit_mode);
    $( "#undo_button" ).click(undo_action);

    $( ".comment_up" ).click(comment_promotion);
    $( "#submit_comment" ).click(function( event ) {
      event.preventDefault();
      submit_comment();
    });
    $('.specline').mouseup(textSelect);
    $('form[name="comment_form"] :button').on("click", function() {
        //alert( "Cancel comment");
        toggle_comment();
    });
    $("#show_comments").click(function(){
          $(".comment").toggle();
    });

    popupmenu();

    /*
    $.contextMenu({
        selector: '.oid', 
        trigger: 'hover',
        autoHide: true,
        delay: 500,
        callback: function(key, options) {
            if (key == "comment") {
                toggle_comment();
                var spec_id = this.context.parentNode.id;
                $('input[name=spec_id]').val(spec_id);
                $('#spec_text').text(this.context.nextSibling.textContent);
            } else {
            var m = "clicked: " + key;
            window.console && console.log(m) || alert(m); 
            }
        },
        items: {
            "comment": {name: "Comment", icon: "comment"},
            "edit": {name: "Edit", icon: "edit"},
            "delete": {name: "Delete", icon: "delete"},
        }
    });
    */

});

