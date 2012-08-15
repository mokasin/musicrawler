"use strict"

soundManager.setup({
	url: "/assets/js/SoundManager2/",
	useHTML5Audio: true,
	debugMode: true,
})

format = (str, obj) ->
	return str.replace(/{(\w+)}/g, (match, id) ->
		return if typeof(obj[id]) == "undefined" then match else obj[id]
	)

# define vlist methods
list = (o) ->
	elem = $(o.context)

	o.clear = ->
		elem.find("ul li").remove()

	o.update = (parentId) ->
		o.clear()

		query = o.url.replace(/%/g, parentId)

		$.getJSON(query, (data) ->
			$.each(data, (key, val) ->
				elem.find("ul").append(
					format(o.template, val)
				)
			)
		)

	o.select = (item) ->
		elem.find("li.active").removeClass("active")
		if item != undefined
			item.parent("li").toggleClass("active", true)

	o.resize = ->
		elem.height(($(window).height() - elem.offset().top))

	return o

# define a case insensitve :contains selector
$.expr[":"].Contains = $.expr.createPseudo((arg) ->
	return (elem) ->
		return $(elem).text().toUpperCase().indexOf(arg.toUpperCase()) >= 0
)

playableList = (o) ->
	that = list(o)

	s = null

	decorate = (item) ->
		that.select(item)

		if s != null
			if s.paused
				item.prepend('<i class="icon-pause"></i>')
			else
				item.prepend('<i class="icon-play"></i>')

	undecorate = (item) ->
		item.parent().find("i").remove()

	that.playPause = (item) ->
		id = item.attr("id")

		if s != null
			oldId = s.id

		if id != oldId
			if s != null
				s.destruct()
				undecorate($("." + oldId, that.context))

			s = soundManager.createSound({
				id: id,
				url: item.attr("href")
			})
			s.play()

			decorate(item)
		else
			s.togglePause()
			if s.paused
				s1 = "play"
				s2 = "pause"
			else
				s1 = "pause"
				s2 = "play"

			e = $("i.icon-" + s1, that.context)
			e.removeClass()
			e.addClass("icon-" + s2)

	superUpdate = that.update

	that.update = (parentId) ->
		superUpdate(parentId)
		if s != null
			decorate($("a#" + s.id))

	return that


# enable a filterable list
listFilter = (list, input) ->
	input.change( ->
		filter = input.val()

		if filter
			list.find("li").hide()
			list.find("a:Contains(" + filter + ")").parent().show()
		else
			list.find("li").show()

	).keyup( ->
		# fire the above change event after every letter
		$(this).change()
	)


# start here
$(document).ready ->

	# initialize list objects
	artist = list({
		context: ".vlists .artist",
		url: "/artist.json",
		template: '<li><a href="#" id="artist-{Id}">{Name}</a></li>'
	})

	album = list({
		context: ".vlists .album",
		url: "/artist/%/albums.json",
		template: '<li><a href="#" id="album-{Id}">{Name}</a></li>'
	})

	track = playableList({
		context: ".vlists .track",
		url: "/album/%/tracks.json",
		template: '<li><a href="{Link}" id="track-{Id}">{Title}</a></li>'
	})

	artist.update()


	# EVENTS

	listFilter($(".vlists .artist ul"), $("form#artist-form input"))

	# fit lists to window
	$(window).resize ->
		artist.resize()
		album.resize()
		track.resize()

	# on click

	# artist
	$(".vlists .artist").on("click", "a", ->
		album.update($(this).attr("id").split("-")[1])
		track.clear()

		artist.select($(this))
	)

	# album
	$(".vlists .album").on("click", "a", ->
		track.update($(this).attr("id").split("-")[1])
		album.select($(this))
	)

	# track


	$(".vlists .track").on("click", "a", (e) ->
		e.preventDefault()

		track.playPause($(this))
	)