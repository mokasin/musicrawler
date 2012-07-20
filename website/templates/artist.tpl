{{define "content"}}
<a class="btn" href="javascript:history.back()">Back to artists</a>

<div class="album-table">
	<table class="table table-condensed table-striped">
		<thead>
			<tr>
				<th>Album</th>
			</tr>
		</thead>
		<tbody>
			{{range .Albums}}
				<tr>
					<td>
						<a href="{{.Path}}">
							{{.Album.Name}}
						</a>
					</td>
				</tr>
			{{end}}
		</tbody>
	</table>
</div>
{{end}}
