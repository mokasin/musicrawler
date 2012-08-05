{{define "content"}}
<style type="text/css">
	html { min-height: 100%; }
</style>

<div class="vlists">
	<div>
		<ul class="nav nav-list">
		{{range .Artists}}
			<li><a href="#">{{.Name}}</a></li>
		{{end}}
		</ul>
	</div>
</div>

{{end}}
