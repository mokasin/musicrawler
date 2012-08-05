{{define "content"}}
<a href="{{.Page.BackLink}}" class="btn">
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
						<a href="{{.Link}}" class="js-pjax">
							{{.Name}}
						</a>
					</td>
				</tr>
			{{else}}
				<tr><td>Artist has no albums.</td></tr>
			{{end}}
		</tbody>
	</table>
</div>
{{end}}
