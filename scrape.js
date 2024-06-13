(function scrape() {
  // https://j.hn/html-table-to-json/
  function tableToJson(table) {
    var data = [], headers = [];
    // first row needs to be headers
    for (var i = 0; i < table.rows[0].cells.length; i++) {
      headers[i] = table.rows[0].cells[i].innerHTML.toLowerCase().replace(
        / /gi,
        "",
      );
    }
    // go through cells
    for (var i = 1; i < table.rows.length; i++) {
      var tableRow = table.rows[i];
      var rowData = {};
      for (var j = 0; j < tableRow.cells.length; j++) {
        rowData[headers[j]] = tableRow.cells[j].innerHTML;
      }
      data.push(rowData);
    }
    return data;
  }

  function download(filename, text) {
    const pom = document.createElement("a");
    pom.setAttribute(
      "href",
      "data:text/json;charset=utf-8," + encodeURIComponent(text),
    );
    pom.setAttribute("download", filename);
    pom.click();
    pom.remove();
  }
  // So deep
  const tbody = document.querySelector("#widgetFrame911")
    .contentDocument.querySelector("#tctPollSystem")
    .contentDocument.querySelector("#AppMainMenu")
    .contentDocument.querySelector("#workspace")
    .contentDocument.querySelector("table#dgShiftDetail tbody");
  download("jb.json", JSON.stringify(tableToJson(tbody)));
})();
