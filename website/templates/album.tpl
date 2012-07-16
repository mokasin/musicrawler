{{define "content"}}

<link href="/assets/widgets/360-player/360player.css" rel="stylesheet" />

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
							<a href="{{.Path}}" title="Play"></a>
						</div>
					</td>
					<td>{{.Track.Artist}}</td>
					<td>{{.Track.Album}}</td>
					<td><a href="{{.Path}}">{{.Track.Title}}</a></td>
					<td>{{.Track.Tracknumber}}</td>
					<td>{{.Track.Year}}</td>
					<td>{{.Track.LengthString}}</td>
				</tr>
			{{end}}
		</tbody>
	</table>
</div>

<!--/ Placed at the end of the document so the pages load faster -->
<script src="/assets/js/SoundManager2/soundmanager2-nodebug-jsmin.js"></script> 
<script src="/assets/js/soundmanager-settings.js"></script>

<script src="/assets/widgets/360-player/script/berniecode-animator.js"></script>
<script src="/assets/widgets/360-player/script/360player.js"></script>
{{end}}
