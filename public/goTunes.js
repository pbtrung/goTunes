// Format times as minutes and seconds.
var timeFormat = function(secs) {
    if (secs == undefined || isNaN(secs)) {
        return "0:00";
    }
    secs = Math.round(secs);
    var mins = "" + Math.round(secs / 60);
    secs = "" + secs % 60;
    if (secs.length < 2) {
        secs = "0" + secs;
    }
    return mins + ":" + secs;
};

var itemSearchUrl = "https://trungpham.cloudant.com/music/_design/item_search/_search/items?q=";
var username = "waskothatedgelefattilled";
var password = "44fdf9671744a563b60bfe6b9833a37d75e2933c";

$(function() {

    // Model
    var Item = Backbone.Model.extend({
    });
    var Items = Backbone.Collection.extend({
        model: Item
    });

    var Router = Backbone.Router.extend({
        routes: {
            "search/:query": "itemQuery"
        },
        itemQuery: function(query) {
            var queryUrl = itemSearchUrl + encodeURIComponent($("#query").val());
            $.ajax({
                type: "GET",
                url: queryUrl,
                dataType: "json",
                headers: {
                    Authorization: "Basic " + btoa(username + ":" + password)
                },
                success: function(data) {
                    var items = [];
                    for (var i = 0; i < data["rows"].length; i++) {
                        items.push(data["rows"][i]["fields"]);
                        items[i]["length"] = timeFormat(items[i]["length"]);
                        items[i]["bitrate"] = Math.round(items[i]["bitrate"] / 1000);
                    }  
                    var models = _.map(items, function(d) {
                        return new Item(d);
                    });
                    var results = new Items(models);
                    app.showResults(results);
                }
            });
        }
    });
    var router = new Router();
    Backbone.history.start();   

    var AppView = Backbone.View.extend({
        el: $("body"),
        events: {
            "submit #queryForm": "routeResults",
            "click #pause": "pauseTrack",
            "click #download": "downloadTrack",
            "click #play": "playTrack"
        },

        pauseTrack: function(ev) {
            ev.preventDefault();
            $("#player").get(0).pause();
        },

        downloadTrack: function(ev) {
            ev.preventDefault();
            var download = $(ev.currentTarget);
            var hiddenDownload = download.parent().find("#hiddenDownload");
            var fileID = download.data("fileid");

            getItemURL(fileID, function(itemURL) {
                hiddenDownload.attr("href", itemURL);
                hiddenDownload.attr("download", itemURL);
                hiddenDownload.get(0).click();
            });
        },

        playTrack: function(ev) {
            ev.preventDefault();
            var play = $(ev.currentTarget);
            var fileID = play.data("fileid");
            getItemURL(fileID, function(itemURL) {
                var player = document.getElementById("player");
                if (player.src !== itemURL) {
                    $("#player").attr("src", itemURL);
                    $("#player").get(0).play();
                } else {
                    $("#player").get(0).play();
                }
            });
        },

        routeResults: function(ev) {
            ev.preventDefault();
            router.navigate("search/" + encodeURIComponent($("#query").val()), { trigger: true });
        },
        showResults: function(results) {
            var source = $("#template").html();
            var template = Handlebars.compile(source);
            var html;
            $("#results").empty();
            results.each(function(result) {
                html = template(result.toJSON());
                $("#results").append(html);
            });

            showHideInfo();
            showHideLyrics();
            
            var options = {
                valueNames: ["searchSrc1", "searchSrc2"],
                page: 5,
                pagination: true
            };
            var resultList = new List("content", options);
        }
    });

    var app = new AppView();
});

function showHideInfo() {
    $("#controls #info").click(function(ev) {
        ev.preventDefault();
        var fileID = $(this).data("fileid");
        var dl = $("#dl-" + fileID);
        if(dl.hasClass("hidden")) {
            $("#results .dl-horizontal").addClass("hidden");
            $("#results .lyrics").addClass("hidden");
            dl.removeClass("hidden");
            if($("#album-art-" + fileID).length) {
                var albumart = $("#album-art-" + fileID);
                getItemURL(albumart.data("albumartid"), function(itemURL) {
                    albumart.attr("src", itemURL);
                });
            }
        } else {
            $("#results .dl-horizontal").addClass("hidden");
        }
    });
}

function showHideLyrics() {
    $("#controls #lyrics").click(function(ev) {
        ev.preventDefault();
        var elId = "#lyrics-" + $(this).data("fileid");
        var lyrics = $(elId);
        if(lyrics.hasClass("hidden")) {
            $("#results .lyrics").addClass("hidden");
            $("#results .dl-horizontal").addClass("hidden");
            lyrics.removeClass("hidden");
        } else {
            $("#results .lyrics").addClass("hidden");
        }
    });
}

function getItemURL(fileID, callback) {
    if(localStorage.getItem("gdrToken") === null) {
        getItemURLFromGDR(fileID, callback);
    } else {
        var gdrToken = JSON.parse(localStorage.getItem("gdrToken"));
        var now = new Date();
        var tokenDate = new Date(gdrToken.Expiration);
        if(now > tokenDate) {
            getItemURLFromGDR(fileID, callback);
        } else {
            getItemURLFromGDRWithToken(fileID, gdrToken, callback); 
        }
    }
}

function getItemURLFromGDR(fileID, callback) {
    $.ajax({
        type: "GET",
        url: "/gdrToken",
        dataType: "json",
        success: function(gdrToken) {
            localStorage.setItem("gdrToken", JSON.stringify(gdrToken));
            getItemURLFromGDRWithToken(fileID, gdrToken, callback);
        }
    });
}

function getItemURLFromGDRWithToken(fileID, gdrToken, callback) {
    $.ajax({
        type: "GET",
        url: "https://www.googleapis.com/drive/v2/files/" + fileID,
        headers: {
            Authorization: "Bearer " + gdrToken.AccessToken
        },
        dataType: "json",
        success: function(item) {
            var itemURL = item.downloadUrl + "&access_token=" + encodeURIComponent(gdrToken.AccessToken);
            callback(itemURL);
        }
    });
}
