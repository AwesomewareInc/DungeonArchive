window.onload = function() {
	queries = document.getElementById("queries");

	// Plus button
	document.getElementById("add-character").addEventListener("click",function() {
		// Get the queries.
		children = queries.childNodes;
		// Get the last query
		lastQuery = document.getElementById("staple");
		// Add a query before that last query
		span = document.createElement("span");
		span.innerHTML = "<input type='text' placeholder='Character'>";
		// Add on some text based on how many queries we have.
		if(children.length-8 > 1) {
			span.innerHTML += " and";
		} else {
			span.innerHTML += " interacting with";
		}
		span.className = "query";
		queries.insertBefore(span, children[5]);
	})

	// Minus button
	document.getElementById("remove-character").addEventListener("click",function() {
		if(queries.childNodes.length-9 == 0) {return}
		queries.childNodes[5].remove();
	})

	document.getElementById("submit").addEventListener("click",function() {
		text = "";
		// we avoid just doing "getElementsByClassName" since some very old browsers don't support that.
		children = queries.childNodes;
		for (q in children) {
			child = children[q]
			if(typeof child != "object") {continue}
			if(child.className != "query") {continue}
			input = child.getElementsByTagName("input")[0];
			text += input.value+",";
		}
		url = window.location.href
		url = url.replace("/search?","",1)
		url = url.replace("/search","",1)
		window.location.href = url+"/results?search="+text;
	})
}
