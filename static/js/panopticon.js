/**
 * Created by oli on 2/11/17.
 */
$(document).ready(function () {

    var healthReq = $.ajax({
        url: "/api/v1/state/current",
        method: "GET"
    }).done(function(data){
        $("#current-state").text(JSON.stringify(data.current));
        $("#last-state").text(JSON.stringify(data.last));
        if (data.message == "") {
            $("#error-message-label").hide();
            $("#error-message").hide();
        } else {
            $("#error-message-label").hide();
            $("#error-message").text(JSON.stringify(data.message)).show();
        }
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var consulReq = $.ajax({
        url: "/api/v1/consul/health",
        method: "GET"
    }).done(function(data){
        $("#consul-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var glusterReq = $.ajax({
        url: "/api/v1/gluster/health",
        method: "GET"
    }).done(function(data){
        $("#gluster-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var weaveReq = $.ajax({
        url: "/api/v1/weave/health",
        method: "GET"
    }).done(function(data){
        $("#weave-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

});