"use strict"

soundManager.setup({
	url: "/assets/js/SoundManager2/",
	useHTML5Audio: true,
	debugMode: true,
})

###############################################################################
# SoundManager manager
###############################################################################
sm = ->
	that = {}

	that.sm = null

	meta = null

	that.playPause = (m) ->
		if m == undefined and meta == null
			throw "No track was played yet."
		else if m == undefined or m == meta
			that.sm.togglePause()
		else
			meta = m

			if that.sm != null
				that.sm.destruct()

			that.sm = soundManager.createSound({
				id: m.Id,
				url: m.Link
			})

			that.sm.play()

		# call event
		$(document).trigger('playpause', [that.sm.paused, meta])

	that.stop = ->
		that.sm.stop()
		$(document).trigger('stop', [meta])

	return that



###############################################################################
# define vlist methods
###############################################################################

format = (str, obj) ->
	return str.replace(/{(\w+)}/g, (match, id) ->
		return if typeof(obj[id]) == "undefined" then match else obj[id]
	)

list = (o) ->
	o.elem = $(o.context)

	o.selectedId

	o.clear = ->
		o.elem.find("ul li").remove()

	o.update = (parentId) ->
		o.clear()
		o.meta = []

		query = o.url.replace(/%/g, parentId)
		list = ""

		$.getJSON(query, (data) ->
			$.each(data, (key, val) ->
				o.meta[val.Id] = val
				list +=	format(o.template, val)
			)

			o.elem.find("ul").append(list)
			o.elem.trigger("updatedList")
		)


	o.select = (item) ->
		o.elem.find("li.active").removeClass("active")
		if item != undefined
			item.parent("li").toggleClass("active", true)
		o.selectedId = item.attr("id")

	o.resize = ->
		o.elem.height(($(window).height() - o.elem.offset().top))

	return o



###############################################################################
# extend normal list -> track list
###############################################################################
playableList = (o) ->
	that = list(o)

	decorate = (item) ->
		that.select(item)

		if s.sm.paused
			item.prepend('<i class="icon-pause"></i>')
		else
			item.prepend('<i class="icon-play"></i>')

	undecorate = (item) ->
		item.parent().find("i").remove()

	that.stop = ->
		undecorate($("a#" + that.selectedId))

	that.playPause = (item) ->
		if item == undefined
			item = $("a#track-" + s.sm.id, that.context)

		undecorate($("a#" + that.selectedId))
		decorate(item)

	# hook into event to highlight the currently played track
	that.elem.bind("updatedList", ->
		if s.sm != null
			decorate($("a#track-" + s.sm.id))
	)

	return that


###############################################################################
# connect list with input to filter it
###############################################################################


# define a case insensitve :contains selector
$.expr[":"].Contains = $.expr.createPseudo((arg) ->
	return (elem) ->
		return $(elem).text().toUpperCase().indexOf(arg.toUpperCase()) >= 0
)

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





###############################################################################
# start here
#

# define global SoundManager-manager
s = sm()
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

	resize = ->
		artist.resize()
		album.resize()
		track.resize()

	resize()

	##################################
	#             EVENTS             #
	##################################

	listFilter($(".vlists .artist ul"), $("form#artist-form input"))
	listFilter($(".vlists .album ul"), $("form#album-form input"))
	listFilter($(".vlists .track ul"), $("form#track-form input"))

	########################
	# fit lists to window
	########################
	$(window).resize ->
		resize()

	########################
	# on click
	########################

	# artist list
	$(".vlists .artist").on("click", "a", ->
		album.update($(this).attr("id").split("-")[1])
		track.clear()

		artist.select($(this))
	)

	# album list
	$(".vlists .album").on("click", "a", ->
		track.update($(this).attr("id").split("-")[1])
		album.select($(this))
	)

	# track list
	$(".vlists .track").on("click", "a", (e) ->
		e.preventDefault()

		mid = $(this).attr("id").split("-")[1];

		s.playPause(track.meta[mid])
	)


	#########################
	# global events
	#########################

	elemNowPlaying = $(".now-playing")
	elemNowPlaying.css("visibility", "hidden")

	# play, stop buttons
	elemNowPlaying.find(".btn-play").on("click", ->
		s.playPause()
	)

	elemNowPlaying.find(".btn-stop").on("click", ->
		s.stop()
	)

	$(document).bind('playpause', (e, paused, meta) ->
		elemNowPlaying.css("visibility", "visible")

		elemNowPlaying.find(".track-artist").text(meta.Artist)
		elemNowPlaying.find(".track-album").text(meta.Album)
		elemNowPlaying.find(".track-title").text(meta.Title)
		elemNowPlaying.find(".track-length").text(meta.Length)

		track.playPause($("a#track-" + meta.Id))

		btn = elemNowPlaying.find(".btn-play i")
		btn.removeClass()
		if paused
			btn.addClass("icon-play")
		else
			btn.addClass("icon-pause")
	)

	$(document).bind('stop', (e, meta) ->
		track.stop()
		btn = elemNowPlaying.find(".btn-play i")
		btn.removeClass()
		btn.addClass("icon-play")
	)