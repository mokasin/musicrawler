soundManager.url = 'assets/soundmanager/swf/';
soundManager.flashVersion = 9; // optional: shiny features (default = 8)
soundManager.useHTML5Audio = true;
soundManager.preferFlash = true;
soundManager.debugMode = true;
soundManager.audioFormats = {                                           
	'mp3': {                                                              
		'type': ['audio/mpeg; codecs="mp3"', 'audio/mpeg', 'audio/mp3',     
		'audio/MPA', 'audio/mpa-robust'],                          
		'required': false                                                   
	},                                                                    
	'ogg': {                                                              
		'type': ['audio/ogg; codecs=vorbis'],                               
		'required': false                                                   
	}                                                                     
};                                                                      
