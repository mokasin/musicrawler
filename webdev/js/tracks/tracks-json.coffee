soundManager.setup({
  url: '/assets/js/SoundManager2/',
		useHTML5Audio: true,
		debugMode: true,
})


format = (val, str) ->
	return str.replace(/{(\w+)}/g, (match, id) ->
		return if typeof(val[id]) == 'undefined' then match else val[id]
	)


list = (o) ->
	elem = $(o.context)

	o.clear = ->
		elem.find('ul li').remove()

	o.update = (parentId) ->
		o.clear()

		query = o.url.replace(/%/g, parentId)

		$.getJSON(query, (data) ->
			$.each(data, (key, val) ->
				elem.find('ul').append(
					format(val, o.template)
				)
			)
		)

	o.select = (item) ->
		elem.find('li.active').removeClass('active')
		item.parent('li').toggleClass('active', true)

	o.resize = ->
		elem.height(($(window).height() - elem.offset().top))

	return o


$(document).ready ->

	# initialize list objects
	artist = list({
		name: 'artist',
		context: '.vlists .artist',
		url: '/artist.json',
		template: '<li><a href="#" class="artist-{Id}">{Name}</a></li>'
	})

	album = list({
		name: 'album',
		context: '.vlists .album',
		url: '/artist/%/albums.json',
		template: '<li><a href="#" class="album-{Id}">{Name}</a></li>'
	})

	track = list({
		name: 'track',
		context: '.vlists .track',
		url: '/album/%/tracks.json',
		template: '<li><a href="{Link}" class="track-{Id}">{Title}</a></li>'
	})

	artist.update()

	# fill with first artist
	album.update(1)

	# EVENTS

	# fit lists to window
	$(window).resize ->
		artist.resize()
		album.resize()
		track:w.resize()

	# on click

	# artist
	$('.vlists .artist').on('click', 'a', ->
		album.update($(this).attr('class').split('-')[1])
		artist.select($(this))
	)

	# album
	$('.vlists .album').on('click', 'a', ->
		track.update($(this).attr('class').split('-')[1])
		album.select($(this))
	)

	# track
	s = null

	$('.vlists .track').on('click', 'a', (e) ->
		e.preventDefault()

		s.destruct() if s != null
		s = soundManager.createSound(
			{id: $(this).attr('class'), url:$(this).attr('href')})
		s.play()

		track.select($(this))
	)