package warewulfd

// TODO: move to separate file?
const nodeStatusHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<style>
table {
  font-family: arial, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

td, th {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}

tr:nth-child(even) {
  background-color: #dddddd;
}
</style>
</head>
<body>
    <h1>{{.PageTitle}}</h1>
    <table>
      <tr>
        <th>Node</th>
        <th>Cluster</th>
        <th>last seen (s)</th>
      </tr>
        {{range .HtmlBody}}
            <tr>
                <td>{{.Node}}</td>
                <td>{{.Cluster}}</td>
                <td>{{.LastSeen}}</td>
            </tr>
        {{end}}
    </table>
</body>
</html>`
