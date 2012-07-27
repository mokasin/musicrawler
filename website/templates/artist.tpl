{{define "content"}}
<a class="btn" href="javascript:history.back()">
	<i class="icon-chevron-left"></i> Back to artists
</a>

<h1 class="artist-title">{{.Artist.Name}}</h1>

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
						<a href="{{.Link}}">
							{{.Name}}
						</a>
					</td>
				</tr>
			{{end}}
		</tbody>
	</table>
</div>
{{end}}
