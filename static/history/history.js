var fetch = function (event) {
  const container = $("#historyContainer");

  jQuery.ajax({
    url: "/list",
    type: "GET",
    contentType: "application/json; charset=utf-8",
    dataType: "json",
    success: function (response) {

      // Find a <table> element with id="myTable":
      var table = document.getElementById("historyTable");


      response.forEach(element => {

        var row = table.insertRow();
        var uid = row.insertCell(0);
        var link = row.insertCell(1);
        var download = row.insertCell(2);
        var size = row.insertCell(3);

        uid.innerHTML = element.uid;
        link.innerHTML = '<a href="/download/' + element.uid + '">' + element.name + '</a>';
        download.innerHTML = '<a href="/download/' + element.uid + '?download=true">#</a>';
        size.innerHTML = humanFileSize(element.size, true);

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