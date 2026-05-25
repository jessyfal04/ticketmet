// Open a concert by id
function openConcert(id) {
	return api("GET", "/api/concerts/" + id)
		.done(function (response) {
			selectedConcert = response;
			renderConcert();
			if (user) {
				loadProfile();
			}
			showView("detailView");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Concert not found."));
		});
}

// Close the concert view
function closeConcert() {
	selectedConcert = null;
	renderConcert();
	showView("searchView");
}

// Share the current concert link
function shareConcert() {
	if (!selectedConcert) return;

	let link = new URL(window.location.href);
	link.search = "";
	link.searchParams.set("concert", itemID(selectedConcert));

	copyTextToClipboard(link.toString())
		.then(function () {
			setMessages("info", "Concert link copied.");
		})
		.catch(function () {
			setMessages("danger", "Unable to copy concert link.");
		});
}

// Copy text to the clipboard
function copyTextToClipboard(text) {
	if (navigator.clipboard && navigator.clipboard.writeText) {
		return navigator.clipboard.writeText(text);
	}

	window.prompt("Copy this link", text);
	return Promise.resolve();
}

// Render the concert details
function renderConcert() {
	if (!selectedConcert) {
		$("#noConcertText").show();
		$("#concertDetails").hide();
		$("#concertDetails").removeClass("ticketmet-concert-detail--expired");
		$("#detailDateStatus").html("");
		$("#detailSaleStatus").html("");
		$("#detailSeatmap").hide();
		renderConcertActions();
		return;
	}

	$("#noConcertText").hide();
	$("#concertDetails").show();
	$("#concertDetails").toggleClass("ticketmet-concert-detail--expired", isExpiredConcert(selectedConcert));
	$("#detailName").text(selectedConcert.Name);
	$("#detailDate").text(formatDate(selectedConcert.Date));
	$("#detailSale").text(formatDate(selectedConcert.SaleStartDateTime) || "Not provided");
	$("#detailVenue").text(selectedConcert.VenueName || "Unknown venue");
	$("#detailArtist").text(selectedConcert.ArtistName || "Unknown artist");
	$("#detailURL").attr("href", selectedConcert.URL || "#");
	$("#detailDateStatus").html(isExpiredConcert(selectedConcert)
		? `<span class="tag is-dark is-light">Past event</span>`
		: `<span class="tag is-success is-light">Upcoming</span>`);
	$("#detailSaleStatus").html("");
	let sale = field(selectedConcert, ["SaleStartDateTime", "saleStartDateTime"], "");
	if (sale) {
		let saleDate = new Date(sale);
		if (!Number.isNaN(saleDate.getTime()) && saleDate.getTime() > Date.now()) {
			$("#detailSaleStatus").html(`<span class="tag is-info is-light">Sale soon</span>`);
		} else if (!Number.isNaN(saleDate.getTime())) {
			$("#detailSaleStatus").html(`<span class="tag is-warning is-light">Sale open</span>`);
		}
	}
	if (selectedConcert.SeatmapURL) {
		$("#detailSeatmap").show().attr("href", selectedConcert.SeatmapURL);
	} else {
		$("#detailSeatmap").hide();
	}

	$("#detailPhoto").attr("src", concertPhoto(selectedConcert));

	renderConcertActions();
	loadConcertFeatures();
}

// Load the SNS and WTB/WTS blocks
function loadConcertFeatures() {
	if (!selectedConcert) {
		$("#detailSetlist").html("");
		$("#detailSNS").html("");
		$("#detailWT").html("");
		return;
	}

	let id = encodeURIComponent(itemID(selectedConcert));

	api("GET", "/api/setlist/" + id)
		.done(function (response) {
			let songs = response.Songs || [];
			if (songs.length == 0) {
				$("#detailSetlist").html(`<span class="has-text-grey">No setlist available.</span>`);
				return;
			}
			let html = "<ol>";
			for (let i = 0; i < songs.length; i = i + 1) {
				html += `<li>${escapeHTML(songs[i])}</li>`;
			}
			html += "</ol>";
			$("#detailSetlist").html(html);
		})
		.fail(function (xhr) {
			$("#detailSetlist").html(escapeHTML(errorText(xhr, "Unable to load setlist.")));
		});

	api("GET", "/api/favorites/" + id)
		.done(function (response) {
			renderTags("#detailSNS", response.SNS || [], "No SNS yet.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			$("#detailSNS").html(escapeHTML(errorText(xhr, "Unable to load SNS.")));
		});

	api("GET", "/api/wt/" + id)
		.done(function (response) {
			renderWT(response);
			renderConcertActions();
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			$("#detailWT").html(escapeHTML(errorText(xhr, "Unable to load WTB/WTS.")));
		});
}

// Render the WTB/WTS section
function renderWT(response) {
	let wtb = response.WTB || [];
	let wts = response.WTS || [];
	let wtbCount = field(response, ["WTBCount", "wtbCount"], wtb.length);
	let wtsCount = field(response, ["WTSCount", "wtsCount"], wts.length);
	let expired = selectedConcert ? isExpiredConcert(selectedConcert) : false;
	let note = expired ? `<p class="has-text-grey mb-2">Past event, trade is disabled.</p>` : "";
	let html = `
		${note}
		<p><strong>WTB:</strong> ${escapeHTML(wtbCount)}</p>
		<div class="tags mb-2">${tagsHTML(wtb, "No WTB SNS.")}</div>
		<p><strong>WTS:</strong> ${escapeHTML(wtsCount)}</p>
		<div class="tags">${tagsHTML(wts, "No WTS SNS.")}</div>`;
	$("#detailWT").html(html);
}

// Render a tag list into the target node
function renderTags(target, values, emptyText) {
	$(target).html(tagsHTML(values, emptyText));
}

// Build the tag list HTML
function tagsHTML(values, emptyText) {
	if (values.length == 0) {
		return `<span class="tag is-light">${escapeHTML(emptyText)}</span>`;
	}
	let html = "";
	for (let i = 0; i < values.length; i = i + 1) {
		html += `<span class="tag is-info is-light">${escapeHTML(values[i])}</span>`;
	}
	return html;
}

// Add the current concert to favorites
function addFavorite() {
	setFavorite("POST", "Favorite added.");
}

// Remove the current concert from favorites
function deleteFavorite() {
	setFavorite("DELETE", "Favorite removed.");
}

// Send a favorite update to the API
function setFavorite(method, message) {
	if (!selectedConcert) return;
	api(method, "/api/favorites/" + encodeURIComponent(itemID(selectedConcert)))
		.done(function () {
			loadProfile();
			loadConcertFeatures();
			setMessages("success", message);
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Favorite update failed."));
		});
}

// Add a WTB or WTS mark
function addWT(type) {
	setWT("POST", type, String(type).toUpperCase() + " added.");
}

// Remove a WTB or WTS mark
function deleteWT(type) {
	setWT("DELETE", type, String(type).toUpperCase() + " removed.");
}

// Send a WTB or WTS update to the API
function setWT(method, type, message) {
	if (!selectedConcert) return;
	api(method, "/api/wt/" + encodeURIComponent(itemID(selectedConcert)) + "?type=" + encodeURIComponent(type))
		.done(function () {
			loadProfile();
			loadConcertFeatures();
			setMessages("success", message);
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "WTB/WTS update failed."));
		});
}

// Render the concert action buttons
function renderConcertActions() {
	let target = document.getElementById("concertActions");
	if (!target || !window.React || !window.ReactDOM) return;
	if (!selectedConcert) {
		ReactDOM.render(null, target);
		return;
	}

	let concertID = String(itemID(selectedConcert));
	let expired = isExpiredConcert(selectedConcert);
	let favorite = hasFavorite(concertID);
	let mineWTB = hasWT(concertID, "wtb");
	let mineWTS = hasWT(concertID, "wts");

	ReactDOM.render(
		React.createElement(ConcertActions, {
			connected: !!user,
			expired: expired,
			favorite: favorite,
			mineWTB: mineWTB,
			mineWTS: mineWTS,
			onAddFavorite: addFavorite,
			onDeleteFavorite: deleteFavorite,
			onAddWTB: function () {
				addWT("wtb");
			},
			onDeleteWTB: function () {
				deleteWT("wtb");
			},
			onAddWTS: function () {
				addWT("wts");
			},
			onDeleteWTS: function () {
				deleteWT("wts");
			},
		}),
		target
	);
}

// Render the concert action buttons
function ConcertActions(props) {
	if (!props.connected) {
		return React.createElement("p", { className: "help mt-2" }, "Sign in to add favorites or WTB/WTS.");
	}

	let favoriteClass = "button " + (props.favorite ? "is-primary is-light" : "is-primary");
	let wtbClass = "button " + (props.mineWTB ? "is-primary is-light" : "is-primary");
	let wtsClass = "button " + (props.mineWTS ? "is-primary is-light" : "is-primary");

	return React.createElement(
		"div",
		{ className: "buttons mt-4" },
		React.createElement(
			"button",
			{
				className: favoriteClass,
				type: "button",
				onClick: props.favorite ? props.onDeleteFavorite : props.onAddFavorite,
				disabled: !props.favorite && props.expired,
			},
			props.favorite ? "Remove favorite" : "Add favorite"
		),
		React.createElement(
			"button",
			{
				className: wtbClass,
				type: "button",
				onClick: props.mineWTB ? props.onDeleteWTB : props.onAddWTB,
				disabled: props.expired,
			},
			props.mineWTB ? "Remove WTB" : "WTB"
		),
		React.createElement(
			"button",
			{
				className: wtsClass,
				type: "button",
				onClick: props.mineWTS ? props.onDeleteWTS : props.onAddWTS,
				disabled: props.expired,
			},
			props.mineWTS ? "Remove WTS" : "WTS"
		)
	);
}

// Check whether the concert is in the past
function isExpiredConcert(concert) {
	if (!concert) return false;
	let value = field(concert, ["Date", "date"], "");
	if (!value) return false;
	let date = new Date(value);
	if (Number.isNaN(date.getTime())) return false;
	return date.getTime() < Date.now();
}

// Read the concert photo or fallback image
function concertPhoto(concert) {
	if (concert && concert.Photos && concert.Photos.length > 0 && concert.Photos[0]) {
		return concert.Photos[0];
	}
	return "img/concert.png";
}

// Build the concert status tags
function renderConcertTags(concert) {
	let tags = [];
	if (isExpiredConcert(concert)) {
		tags.push(React.createElement("span", { className: "tag is-dark is-light", key: "past" }, "Past event"));
	} else {
		tags.push(React.createElement("span", { className: "tag is-success is-light", key: "upcoming" }, "Upcoming"));
	}

	let sale = field(concert, ["SaleStartDateTime", "saleStartDateTime"], "");
	if (sale) {
		let saleDate = new Date(sale);
		if (!Number.isNaN(saleDate.getTime()) && saleDate.getTime() > Date.now()) {
			tags.push(React.createElement("span", { className: "tag is-info is-light", key: "sale-soon" }, "Sale soon"));
		} else if (!Number.isNaN(saleDate.getTime())) {
			tags.push(React.createElement("span", { className: "tag is-warning is-light", key: "sale-open" }, "Sale open"));
		}
	}

	return React.createElement(React.Fragment, null, tags);
}

// Check whether the user already favorited the concert
function hasFavorite(concertID) {
	let favorites = profileData ? profileData.Favorites || [] : [];
	for (let i = 0; i < favorites.length; i = i + 1) {
		if (String(itemID(favorites[i])) == String(concertID)) return true;
	}
	return false;
}

// Check whether the user already marked the concert
function hasWT(concertID, type) {
	let items = profileData ? profileData.WT || [] : [];
	for (let i = 0; i < items.length; i = i + 1) {
		let concert = items[i].Concert || {};
		if (String(itemID(concert)) == String(concertID) && String(items[i].Type) == type) return true;
	}
	return false;
}
