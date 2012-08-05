{{define "content"}}
<a href="{{.Page.BackLink}}" class="btn">
	<i class="icon-chevron-left"></i> Back to artist
</a>

<link href="/assets/widgets/360-player/360player.css" rel="stylesheet" />

<h1 class="album-title">{{.Album.Name}}</h1>

<div class="table-album">
	<table class="table table-condensed table-striped">
		<thead>
			<tr>
				<th></th>
				<th>Artist</th>
				<th>Album</th>
				<th>Title</th>
				<th>Track</th>
				<th>Year</th>
				<th>Length</th>
			</tr>
		</thead>
		<tbody>
			{{range .Tracks}}
				<tr>
					<td>
						<div class="sm2-inline-list ui360">
							<a href="{{.Link}}" title="Play"></a>
						</div>
					</td>
					<td>{{.Artist}}</td>
					<td>{{.Album}}</td>
					<td><a href="{{.Link}}">{{.Title}}</a></td>
					<td>{{.Tracknumber}}</td>
					<td>{{.Year}}</td>
					<td>{{.LengthString}}</td>
				</tr>
			{{else}}
				<tr>
					<td>Album has no tracks.</td>
				</tr>
			{{end}}
		</tbody>
	</table>
</div>

<!--/ Placed at the end of the document so the pages load faster -->
<script src="/assets/js/SoundManager2/soundmanager2.js"></script> 
<script src="/assets/js/soundmanager-settings.js"></script>

<script src="/assets/widgets/360-player/script/berniecode-animator.js"></script>
<script src="/assets/widgets/360-player/script/360player.js"></script>
{{end}}
