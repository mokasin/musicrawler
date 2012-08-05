# $.extend({
# 	getUrlVars: ->
# 		hashes = window.location.href.slice(
# 			window.location.href.indexOf('?') + 1).split('&')
# 		vars = []
# 		
# 		for h in hashes
# 			hash = h.split('=')
# 			vars.push(hash[0])
# 			vars[hash[0]] = hash[1]
# 
# 		return vars
# 	 
# 	getUrlVar: (name) ->
# 		$.getUrlVars()[name]
# })

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
				'<li><a href="#" class="track-"' + val.Id + '">' +
				val.Title +
				'</a></li>'
			)
		)
	)

$(document).ready ->
	#vars = $.getUrlVars()
	
	fillArtists()

	#fill with first artist
	fillAlbums(1)

	$('.vlists .artist').on('click', 'a', ->
		fillAlbums($(this).attr('class').split('-')[1])
	)

	$('.vlists .album').on('click', 'a', ->
		fillTracks($(this).attr('class').split('-')[1])
	)
