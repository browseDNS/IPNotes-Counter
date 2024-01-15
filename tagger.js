// The ipnotes tagger is a simple script that sends a request to the ipnotes.page server.
// You can replicate this script on your own server if you want to self host the tagger instead.

function ipNotesTag() {
    // var url = "http://127.0.0.1:7711/count";
    var url = "https://ipnotes.page/count";
    var xhr = new XMLHttpRequest();
    xhr.open("POST", url);
    xhr.send();
}

ipNotesTag();