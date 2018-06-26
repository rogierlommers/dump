var fetchFiles = function (event) {
  const container = $("#listFilesContainer");

  jQuery.ajax({
    url: "/list-files",
    type: "GET",
    contentType: "application/json; charset=utf-8",
    dataType: "json",
    success: function (response) {

      // Find a <table> element with id="myTable":
      var table = document.getElementById("filesTable");

      response.forEach(element => {

        var row = table.insertRow();
        var uid = row.insertCell(0);
        var link = row.insertCell(1);
        var download = row.insertCell(2);
        var size = row.insertCell(3);

        uid.innerHTML = element.uid;
        link.innerHTML = '<a href="/download/' + element.uid + '?download=false">' + element.name + '</a>';
        download.innerHTML = '<a href="/download/' + element.uid + '?download=true">#</a>';
        size.innerHTML = humanFileSize(element.size, true);

      });

    },

    error: function (e) {
      console.warn(e);
    }

  });
};

var fetchHistory = function (event) {
  const container = $("#listHistoryContainer");

  jQuery.ajax({
    url: "/list-download-history",
    type: "GET",
    contentType: "application/json; charset=utf-8",
    dataType: "json",
    success: function (response) {

      // Find a <table> element with id="myTable":
      var table = document.getElementById("historyTable");

      response.forEach(element => {
        var row = table.insertRow();

        var name = row.insertCell(0);
        var referer = row.insertCell(1);
        var remote_address = row.insertCell(2);
        var timestamp = row.insertCell(3);

        name.innerHTML = element.name;
        referer.innerHTML = element.referer;
        remote_address.innerHTML = element.remote_address;
        timestamp.innerHTML = moment(element.timestamp_download, "YYYY-MM-DDTHH:mm").fromNow(); // 2018-06-26T08:22:30.719825825+02:00

      });

    },

    error: function (e) {
      console.warn(e);
    }

  });
};

function humanFileSize(bytes, si) {
  var thresh = si ? 1000 : 1024;
  if (Math.abs(bytes) < thresh) {
    return bytes + ' B';
  }
  var units = si ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'] : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
  var u = -1;
  do {
    bytes /= thresh;
    ++u;
  } while (Math.abs(bytes) >= thresh && u < units.length - 1);
  return bytes.toFixed(1) + ' ' + units[u];
}