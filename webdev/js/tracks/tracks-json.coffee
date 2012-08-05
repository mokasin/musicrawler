#$.playable('/assets/js/SoundManager2/', {
#		useHTML5Audio: true,
#		debugMode: true,
#	})


soundManager.setup({
  url: '/assets/js/SoundManager2/',
		useHTML5Audio: true,
		debugMode: true,
})


fillArtists = ->
	$('.vlists .artist ul li').remove()

	$.getJSON('/artist.json', (data) ->

		$.each(data, (key, val) ->
			$('.vlists .artist ul').append(
				'<li><a href="#" class="artist-' + val.Id + '">' +
				val.Name +
				'</a></li>'
			)
		)
	)

fillAlbums = (artistId) ->
	$('.vlists .album ul li').remove()
	# track list isn't up to date
	$('.vlists .track ul li').remove()

	$.getJSON('/artist/' + artistId  + '/albums.json', (data) ->
		$.each(data, (key, val) ->
			$('.vlists .album ul').append(
				'<li><a href="#" class="album-' + val.Id + '">' +
				val.Name +
				'</a></li>'
			)
		)
	)

fillTracks = (albumId) ->
	$('.vlists .track ul li').remove()

	$.getJSON('/album/' + albumId  + '/tracks.json', (data) ->
		$.each(data, (key, val) ->
			$('.vlists .track ul').append(
				'<li><a href="'+ val.Link +
				'" class="track-' + val.Id +
				'">' + val.Title +
				'</a></li>'
			)
		)
	)


$(document).ready ->

	resizeVLists = ->
		d =	$('.vlists div div')
		d.height($(window).height() - d.offset().top)


	resizeVLists()

	$(window).resize ->
		resizeVLists()

	fillArtists()

	#fill with first artist
	fillAlbums(1)

	$('.vlists .artist').on('click', 'a', ->
		fillAlbums($(this).attr('class').split('-')[1])

		$('.vlists .artist li.active').removeClass('active')
		$(this).parent('li').toggleClass('active', true)
	)

	$('.vlists .album').on('click', 'a', ->
		fillTracks($(this).attr('class').split('-')[1])

		$('.vlists .album li.active').removeClass('active')
		$(this).parent('li').toggleClass('active', true)
	)

	s = null

	$('.vlists .track').on('click', 'a', (e) ->
		e.preventDefault()

		s.destruct() if s != null
		s = soundManager.createSound(
			{id: $(this).attr('class'), url:$(this).attr('href')})
		s.play()

		$('.vlists .track li.active').removeClass('active')
		$(this).parent('li').toggleClass('active', true)
	)
