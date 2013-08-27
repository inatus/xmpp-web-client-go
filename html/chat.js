const url = "ws://{{.Host}}:{{.Port}}/websocket";

var ws;
var userName;
var logs = new Object();

if ("WebSocket" in window) {
    ws = new WebSocket(url);
} else if ("MozWebSocket" in window) {
    ws = new MozWebSocket(url);
}

$(document).ready(function () {

    $("#login").submit(function () {

        console.log("#login submit called");
        userName = $("#username").val();
        var data = {
            "Type": "login",
            "Data": {
                "UserName": userName,
                "Password": $("#password").val(),
                "Server": $("#server").val()
            }
        }

        console.log("ws connection created");

        ws.send(JSON.stringify(data));

        return false;
    });

});

ws.onmessage = function (event) {
    console.log("Receive message:" + event.data);

    var message = JSON.parse(event.data);
    if (message.Type == "presence") {
       var color
       if (message.Data["Mode"] == null) {
           color = "#808080"; // gray
       } else if (message.Data["Mode"] == "away") {
           color = "#FF0000"; // red
       } else if (message.Data["Mode"] == "available") {
           color = "#0000FF"; // blue
       }
       $(
           "select#userlist option[value='" + message.Data["Remote"] + "']").css("color", color);

    } else if (message.Type == "chat") {
        var remote = $(
            "select#userlist option[value='" + message.Data.Remote + "']")
            .html();
        var text = "<font color=\"red\">(" + remote + ") " + message.Data.Text + "</font><br>";
        if (logs[message.Data.Remote] == undefined) {
            logs[message.Data.Remote] = "";
        }
        logs[message.Data.Remote] += text;
        if ($(":selected").attr("value") == message.Data.Remote) {
            $("div.log").append(text);
        } else {
            $("select#userlist option[value='" + message.Data.Remote + "']")
                .css("background", "#00B9EF");
        }
    } else if (message.Type == "roster") {
        for (var i = 0; i < message.Roster.length; i++) {
            var color
            if (message.Roster[i].Mode == null) {
                color = "#808080"; // gray
            } else if (message.Roster[i].Mode == "away") {
                color = "#FF0000"; // red
            } else if (message.Roster[i].Mode == "available") {
                color = "#0000FF"; // blue
            }
			if (message.Roster[i].Name == "") {
				message.Roster[i].Name = message.Roster[i].Jid;
			}
            $("select#userlist").append(
                $('<option>').html(message.Roster[i].Name).val(
                message.Roster[i].Jid).css("color", color));
        }
    } else if (message.Type == "login") {
        $("body")
            .load(
            "chat.html .container", function () {
            $("#chat")
                .submit(function () {
                var message = $("#message")
                    .val();
                var data = {
                    "Type": "chat",
                    "Data": {
                        "Remote": $(
                            ":selected")
                            .attr("value"),
                        "Text": message
                    }
                };

                ws.send(JSON.stringify(data));
                console.log("Send message:" + JSON.stringify(data));
                var text = "<font color=\"blue\">(" + userName + ") " + message + "</font><br>";
                if (logs[$(":selected").attr(
                    "value")] == undefined) {
                    logs[$(":selected").attr(
                        "value")] = "";
                }
                logs[$(":selected").attr(
                    "value")] += text;
                $("div.log").append(text);
                $("#message").val("");
                return false;
            });

            $("select#userlist")
                .click(function () {
                console.log(logs[$(":selected")
                    .attr("value")]);
                $(":selected").css(
                    "background", "#FFF");
                if (logs[$(":selected").attr(
                    "value")] != undefined) {
                    $("div.log")
                        .html(
                        logs[$(
                        ":selected")
                        .attr(
                        "value")]);
                } else {
                    $("div.log").html("");
                }
            });
        });
    }
};

ws.onclose = function (event) {
    if (event.wasClean) {
        var closed = "Complete";
    } else {
        var closed = "Incompleted";
    }
    console.log("ws connection disconnected:" + closed + ", code:" + event.code + ", reason:" + event.reason);
    alert("Connection disconnected.");
    window.close();
}

// Close connection
window.onunload = function () {
    var code = 4500;
    var reason = "Client closed";
    ws.close(code, reason);
    console.log("ws connection disposed");
}
