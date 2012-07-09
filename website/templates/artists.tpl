<div class='btn-group' style='margin-bottom: 0.5em'>
	{{range .Pager}}
		{{if .Active}}
			<a class='btn btn-primary' href='#'>{{.Label}}</a>
		{{else}}
			<a class='btn' href='{{.Path}}'>{{.Label}}</a>
		{{end}}
	{{end}}
</div>

<ul class='breadcrumb'>
	{{range .Breadcrumb}}
		{{if .Active}}
			<li class='active'>
				<a href='{{.Path}}'>{{.Label}}</a>
				<span class='divider'>/</span>
			</li>
		{{else}}
			<li>
				<a href='{{.Path}}'>{{.Label}}</a>
				<span class='divider'>/</span>
			</li>
		{{end}}
	{{end}}
</ul>

<div class='artist-table'>
	<table class='table table-condensed table-striped'>
		<thead>
			<tr>
				<th>Artist</th>
			</tr>
		</thead>
		<tbody>
			{{range .Artists}}
				<tr>
					<td>
						<a href='{{.URL}}'>
							{{.Name}}
						</a>
					</td>
				</tr>
			{{end}}
		</tbody>
	</table>
</div>
