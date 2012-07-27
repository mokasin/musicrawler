{{define "content"}}
<h1>Albums</h1>

<div class="table-albums">
	<table class="table table-condensed table-striped">
		<thead>
			<tr>
				<th>Album</th>
			</tr>
		</thead>
		<tbody>
			{{range .Albums}}
				<tr>
					<td><a href="{{.Path}}">{{.Album.Name}}</a></td>
				</tr>
			{{end}}
		</tbody>
	</table>
</div>
{{end}}
