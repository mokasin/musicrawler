<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
	<head>
		<meta charset='utf-8' />
		<title>{{.page.Title}}</title>
		<link href='/assets/css/bootstrap.css' rel='stylesheet' />
		<link href='/assets/css/responsive.css' rel='stylesheet' />
		<style>
			body { padding-top: 60px }
		</style>
		<!-- HTML5 shim, for IE6-8 support of HTML5 elements -->
		<!--[if IE 9]>
			<script src='http://html5shim.googlecode.com/svn/trunk/html5.js'></script>
		<![endif]-->
	</head>
	<body>
		<div class='navbar navbar-fixed-top'>
			<div class='navbar-inner'>
				<div class='container'>
					<a class='btn btn-navbar' data-target='.nav-collapse' data-toggle='collapse'>
						<span class='icon-bar'></span>
						<span class='icon-bar'></span>
						<span class='icon-bar'></span>
					</a>
					<a class='brand' href='#'>musicrawler</a>
					<div class='nav-collapse'>
						<ul class='nav'>
							<li class='active'>
								<a href='#'>Home</a>
							</li>
							<li>
								<a href='#about'>About</a>
							</li>
						</ul>
					</div>
				</div>
			</div>
		</div>
		<div class='container'>
			{{template "content" .content}}
		</div>
	</body>
</html>
