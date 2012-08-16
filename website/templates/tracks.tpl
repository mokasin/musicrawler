{{define "content"}}

<noscript>
<div class="row">
	<div class="span4 offset4">
		Sorry, but without JavaScript enabled, this view is pointless.
	</div>
</div>
</noscript>

<div class="row">
	<div class="span4 offset4 now-playing">
		<h4 class="track-title"></h4>
		by <span class="track-artist"></span>
		from <span class="track-album"></span>

		<div class="btn-group" style="align: center">
			<button class="btn btn-play">
				<i class="icon-pause"></i>
			</button>
			<button class="btn btn-stop">
				<i class="icon-stop"></i>
			</button>
		</div>
	</div>
</div>

<div class="row">
	<div class="vlists span14">
		<div class="row">
			<div class="span4">
				<h3>Artists</h3>

				<form id="artist-form" action="#">
					<div class="input-prepend">
						<span class="add-on"><i class="icon-search"></i></span>
						<input type="search" placeholder="Search...">
					</div>
				</form>

				<div class="artist">
					<ul class="nav nav-list">
					</ul>
				</div>
			</div>

			<div class="span4">
				<h3>Albums</h3>

				<form id="album-form" action="#">
					<div class="input-prepend">
						<span class="add-on"><i class="icon-search"></i></span>
						<input type="search" placeholder="Search...">
					</div>
				</form>

				<div class="album">
					<ul class="nav nav-list">
					</ul>
				</div>
			</div>

			<div class="span4">
				<h3>Tracks</h3>

				<form id="track-form" action="#">
					<div class="input-prepend">
						<span class="add-on"><i class="icon-search"></i></span>
						<input type="search" placeholder="Search...">
					</div>
				</form>

				<div class="track">
					<ul class="nav nav-list">
					</ul>
				</div>
			</div>
		</div>
	</div>
</div>

<script src="/assets/js/jquery-1.8.0.min.js"></script>
<script src="/assets/js/SoundManager2/soundmanager2.js"></script>
<script src="/assets/js/tracks-json.js"></script>
{{end}}
