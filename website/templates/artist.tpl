{{define "content"}}
<ul class="breadcrumb">
	{{range .Breadcrumb}}
		{{if .Active}}
			<li class="active">
				<a href="{{.Path}}">{{.Label}}</a>
				<span class="divider">/</span>
			</li>
		{{else}}
			<li>
				<a href="{{.Path}}">{{.Label}}</a>
				<span class="divider">/</span>
			</li>
		{{end}}
	{{end}}
</ul>

<div class="album-table">
	<table class="table table-condensed table-striped">
		<thead>
			<tr>
				<th>Artist</th>
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
