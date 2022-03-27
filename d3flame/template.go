package d3flame

import (
	// use go embed to load js and css source file
	_ "embed"
)

// D3.js is a JavaScript library for manipulating documents based on data.
// https://github.com/d3/d3

// d3-flame-graph is a D3.js plugin that produces flame graphs from hierarchical data.
// https://github.com/spiermar/d3-flame-graph

// https://cdn.jsdelivr.net/gh/spiermar/d3-flame-graph@2.0.3/dist/d3-flamegraph.css
//go:embed d3-flamegraph.css
var d3Css string

// https://d3js.org/d3.v4.min.js
//go:embed d3.v4.min.js
var d3Js string

// https://cdn.jsdelivr.net/gh/spiermar/d3-flame-graph@2.0.3/dist/d3-flamegraph.min.js
//go:embed d3-flamegraph.min.js
var d3FlameGraphJs string

// https://cdnjs.cloudflare.com/ajax/libs/d3-tip/0.9.1/d3-tip.min.js
//go:embed d3-tip.min.js
var d3TipJs string

// https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css
//go:embed bootstrap.min.css
var bootstrapCSS string

const html = `
<head>
	<title>FlameGraph</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	 <style>
	
		/* Space out content a bit */
		body {
		  padding-top: 20px;
		  padding-bottom: 20px;
		}
	
		/* Custom page header */
		.header {
		  padding-bottom: 20px;
		  padding-right: 15px;
		  padding-left: 15px;
		  border-bottom: 1px solid #e5e5e5;
		}
	
		/* Make the masthead heading the same height as the navigation */
		.header h3 {
		  margin-top: 0;
		  margin-bottom: 0;
		  line-height: 40px;
		}
	
		/* Customize container */
		.container {
		  max-width: 990px;
		}
    </style>

</head>
<body>
	<style type="text/css">{{.D3Css}}</style>
	<style type="text/css">{{.BootstrapCSS}}</style>
	<script type="text/javascript">{{.D3Js}}</script>
	<script type="text/javascript">{{.D3Flame}}</script>
	<script type="text/javascript">{{.D3Tip}}</script>
	<div class="container">
	  <div class="header clearfix">
		<nav>
		  <div class="pull-right">
			<form class="form-inline" id="form">
			  <a class="btn" href="javascript: resetZoom();">Reset zoom</a>
			  <a class="btn" href="javascript: clear();">Clear</a>
			  <div class="form-group">
				<input type="text" class="form-control" id="term">
			  </div>
			  <a class="btn btn-primary" href="javascript: search();">Search</a>
			</form>
		  </div>
		</nav>
		<h3 class="text-muted">d3-flame-graph</h3>
	  </div>
	  <div id="chart">
	  </div>
	  <hr>
	  <div id="details">
	  </div>
	</div>
	<script type="text/javascript">
	   var flameGraph = d3.flamegraph()
		  .width(960)
		  .cellHeight(18)
		  .transitionDuration(750)
		  .minFrameSize(5)
		  .transitionEase(d3.easeCubic)
		  .sort(true)
		  //Example to sort in reverse order
		  //.sort(function(a,b){ return d3.descending(a.name, b.name);})
		  .title("")
		  .onClick(onClick)
		  .differential(false)
		  .selfValue(false);
	
	
		// Example on how to use custom tooltips using d3-tip.
		// var tip = d3.tip()
		//   .direction("s")
		//   .offset([8, 0])
		//   .attr('class', 'd3-flame-graph-tip')
		//   .html(function(d) { return "name: " + d.data.name + ", value: " + d.data.value; });
	
		// flameGraph.tooltip(tip);
	
		var details = document.getElementById("details");
		flameGraph.setDetailsElement(details);
	
		// Example on how to use custom labels
		// var label = function(d) {
		//  return "name: " + d.name + ", value: " + d.value;
		// }
		// flameGraph.label(label);
	
		// Example of how to set fixed chart height
		// flameGraph.height(540);
	
		d3.json("stacks.json", function(error, data) {
		  if (error) return console.warn(error);
		  d3.select("#chart")
			  .datum(data)
			  .call(flameGraph);
		});
	
		document.getElementById("form").addEventListener("submit", function(event){
		  event.preventDefault();
		  search();
		});
	
		function search() {
		  var term = document.getElementById("term").value;
		  flameGraph.search(term);
		}
	
		function clear() {
		  document.getElementById('term').value = '';
		  flameGraph.clear();
		}
	
		function resetZoom() {
		  flameGraph.resetZoom();
		}
	
		function onClick(d) {
		  console.info("Clicked on " + d.data.name);
		}
		</script>
	</script>
</body>
`
