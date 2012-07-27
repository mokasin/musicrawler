{{define "content"}}
<div class="btn-group" style="margin-bottom: 0.5em">
	{{range .Pager}}
		{{if .Active}}
			<a class="btn btn-primary" href="{{.Link}}">{{.Label}}</a>
		{{else}}
			<a class="btn" href="{{.Link}}">{{.Label}}</a>
		{{end}}
	{{end}}
</div>

<div class="artist-table">
	<table class="table table-condensed table-striped">
		<thead>
			<tr>
				<th>Artist</th>
			</tr>
		</thead>
		<tbody>
			{{range .Artists}}
				<tr>
					<td>
						<a href="{{.Link}}">
							{{.Name}}
						</a>
					</td>
				</tr>
			{{else}}
				<tr><td>No such artist in database.</td></tr>
			{{end}}
		</tbody>
	</table>
</div>
{{end}}
