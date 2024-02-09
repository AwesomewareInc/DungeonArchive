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
		span.innerHTML = "<input type='text'>";
		// Add the option selector
		span.innerHTML += document.getElementById("options_skeleton").innerHTML;
		span.className = "query";
		queries.insertBefore(span, children[5]);
	})

	// Minus button
	document.getElementById("remove-character").addEventListener("click",function() {
		if(queries.childNodes.length-9 == 0) {return}
		queries.childNodes[5].remove();
	})

	document.getElementById("submit").addEventListener("mouseup",function(e) {
		windowOpenOptions = "_blank";
		if(e.which == 1) {
			windowOpenOptions = "_self";
		}
		text = "";
		// we avoid just doing "getElementsByClassName" since some very old browsers don't support that.
		children = queries.childNodes;
		shouldContinue = true;
		for (q in children) {
			child = children[q]

			if(typeof child != "object") {continue}
			if(child.getElementsByTagName == undefined) {continue}

			// get the value of the input box.
			input = child.getElementsByTagName("input")[0];
			if(input != undefined) {
				input.classList.remove("unselected")
				if(input.value == "") {
					shouldContinue = false;
					input.classList.add("unselected")
				} else {
					text += input.value;
				}
			}
			
			// get the value of the dropdown next to it.
			dropdown = child.getElementsByTagName("select");
			if(dropdown.length <= 0) {continue}
			option = dropdown[0].options[dropdown[0].selectedIndex].value
			dropdown[0].classList.remove("unselected")
			if(option == "") {
				shouldContinue = false;
				dropdown[0].classList.add("unselected")
			} else {
				text += "::"+option
			}
			text += ","
		}
		if(shouldContinue) {
			url = window.location.href
			url = url.replace("/interactionssearch?","",1)
			url = url.replace("/interactionssearch","",1)
			window.open(url+"/interactionsresults?search="+text,windowOpenOptions);
		}

	})
}
