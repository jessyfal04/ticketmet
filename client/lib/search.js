let countryNames = { US: "United States Of America", AD: "Andorra", AI: "Anguilla", AR: "Argentina", AU: "Australia", AT: "Austria", AZ: "Azerbaijan", BS: "Bahamas", BH: "Bahrain", BB: "Barbados", BE: "Belgium", BM: "Bermuda", BR: "Brazil", BG: "Bulgaria", CA: "Canada", CL: "Chile", CN: "China", CO: "Colombia", CR: "Costa Rica", HR: "Croatia", CY: "Cyprus", CZ: "Czech Republic", DK: "Denmark", DO: "Dominican Republic", EC: "Ecuador", EE: "Estonia", FO: "Faroe Islands", FI: "Finland", FR: "France", GE: "Georgia", DE: "Germany", GH: "Ghana", GI: "Gibraltar", GB: "Great Britain", GR: "Greece", HK: "Hong Kong", HU: "Hungary", IS: "Iceland", IN: "India", IE: "Ireland", IL: "Israel", IT: "Italy", JM: "Jamaica", JP: "Japan", KR: "Korea, Republic of", LV: "Latvia", LB: "Lebanon", LT: "Lithuania", LU: "Luxembourg", MY: "Malaysia", MT: "Malta", MX: "Mexico", MC: "Monaco", ME: "Montenegro", MA: "Morocco", NL: "Netherlands", AN: "Netherlands Antilles", NZ: "New Zealand", ND: "Northern Ireland", NO: "Norway", PE: "Peru", PL: "Poland", PT: "Portugal", RO: "Romania", RU: "Russian Federation", LC: "Saint Lucia", SA: "Saudi Arabia", RS: "Serbia", SG: "Singapore", SK: "Slovakia", SI: "Slovenia", ZA: "South Africa", ES: "Spain", SE: "Sweden", CH: "Switzerland", TW: "Taiwan", TH: "Thailand", TT: "Trinidad and Tobago", TR: "Turkey", UA: "Ukraine", AE: "United Arab Emirates", UY: "Uruguay", VE: "Venezuela" };

// Load artists from the API
function loadArtists() {
	api("GET", "/api/artists")
		.done(function (response) {
			artists = response || [];
			renderArtists();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Artist search failed."));
		});
}

// Load venues from the API
function loadVenues() {
	api("GET", "/api/venues")
		.done(function (response) {
			venues = response || [];
			renderVenues();
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Venue search failed."));
		});
}

// Render the artist select
function ArtistResults(props) {
	return React.createElement(
		"div",
		{ className: "select is-fullwidth" },
		React.createElement(
			"select",
			{
				id: "artistResults",
				value: props.value || "",
				onChange: function (event) {
					selectArtist(event.target.value);
				},
			},
			React.createElement("option", { value: "" }, "All artists"),
			(props.items || []).map(function (artist, index) {
				let id = String(itemID(artist));
				return React.createElement("option", { value: id, key: id || String(index) }, itemName(artist) + " (#" + id + ")");
			})
		)
	);
}

// Mount the artist select
function renderArtists() {
	ReactDOM.render(
		React.createElement(ArtistResults, { items: artists, value: selectedArtistID }),
		document.getElementById("artistResultsHere")
	);
	renderAlertBlock();
}

// Render the venue select
function VenueResults(props) {
	return React.createElement(
		"div",
		{ className: "select is-fullwidth" },
		React.createElement(
			"select",
			{
				id: "venueResults",
				value: props.value || "",
				onChange: function (event) {
					selectVenue(event.target.value);
				},
			},
			React.createElement("option", { value: "" }, "All venues"),
			(props.items || []).map(function (venue, index) {
				let id = String(itemID(venue));
				return React.createElement("option", { value: id, key: id || String(index) }, itemName(venue) + " - " + (venue.City || "") + " (#" + id + ")");
			})
		)
	);
}

// Mount the venue select
function renderVenues() {
	ReactDOM.render(
		React.createElement(VenueResults, { items: venues, value: selectedVenueID }),
		document.getElementById("venueResultsHere")
	);
	renderAlertBlock();
	renderCountries();
}

// Update the selected artist
function selectArtist(value) {
	if (value === undefined) value = document.getElementById("artistResults").value;
	selectedArtistID = value || "";
	renderArtists();
	loadConcerts(true);
}

// Update the selected venue
function selectVenue(value) {
	if (value === undefined) value = document.getElementById("venueResults").value;
	selectedVenueID = value || "";
	renderVenues();
	loadConcerts(true);
}

// Mount the country select
function renderCountries() {
	let seen = {};
	let countries = [];
	for (let i = 0; i < venues.length; i = i + 1) {
		let country = field(venues[i], ["Country", "country"], "");
		if (!country || seen[country]) continue;
		seen[country] = true;
		countries.push(country);
	}
	countries.sort(function (left, right) {
		let leftName = countryNames[left] || left;
		let rightName = countryNames[right] || right;
		return leftName.localeCompare(rightName);
	});

	ReactDOM.render(
		React.createElement(CountryResults, { items: countries, value: selectedCountry }),
		document.getElementById("countryResultsHere")
	);
}

// Render the country select
function CountryResults(props) {
	return React.createElement(
		"div",
		{ className: "select is-fullwidth" },
		React.createElement(
			"select",
			{
				id: "countryResults",
				value: props.value || "",
				onChange: function (event) {
					selectCountry(event.target.value);
				},
			},
			React.createElement("option", { value: "" }, "All countries"),
			(props.items || []).map(function (country, index) {
				return React.createElement("option", { value: country, key: country || String(index) }, countryNames[country] || country);
			})
		)
	);
}

// Update the selected country
function selectCountry(value) {
	if (value === undefined) value = document.getElementById("countryResults").value;
	selectedCountry = value || "";
	renderCountries();
	loadConcerts(true);
}

// Update the time filter
function selectStatus() {
	selectedStatus = $("#statusResults").val() || "future";
	loadConcerts(true);
}

// Reset the search filters
function clearSearch() {
	selectedArtistID = "";
	selectedVenueID = "";
	selectedCountry = "";
	selectedStatus = "future";
	renderArtists();
	renderVenues();
	renderCountries();
	$("#statusResults").val(selectedStatus);
	renderAlertBlock();
	loadConcerts(true);
}

// Load concerts from the API
function loadConcerts(reset) {
	if (reset) {
		concertRequestNonce = concertRequestNonce + 1;
		concertPage = 1;
		concertHasMore = true;
		concerts = [];
		concertLoading = false;
		renderConcerts();
	}
	if (concertLoading || !concertHasMore) return;

	let requestNonce = concertRequestNonce;
	let page = concertPage;
	let params = {};
	if (selectedArtistID) params.artistID = selectedArtistID;
	if (selectedVenueID) params.venueID = selectedVenueID;
	params.country = selectedCountry || "all";
	params.status = selectedStatus || "future";
	params.page = page;
	concertLoading = true;

	api("GET", "/api/concerts", params)
		.done(function (response) {
			if (requestNonce != concertRequestNonce) return;
			response = response || [];
			if (response.length == 0) {
				concertHasMore = false;
				return;
			}
			concerts = concerts.concat(response);
			concertPage = page + 1;
			renderConcerts();
		})
		.fail(function (xhr) {
			if (requestNonce != concertRequestNonce) return;
			$("#concertsGrid").html(`<div class="column is-12"><div class="notification is-danger is-light">${escapeHTML(errorText(xhr, "Unable to load concerts."))}</div></div>`);
		})
		.always(function () {
			if (requestNonce == concertRequestNonce) {
				concertLoading = false;
			}
		});
}

// Mount the concert grid
function renderConcerts() {
	ReactDOM.render(React.createElement(ConcertGrid, { items: concerts }), document.getElementById("concertsGrid"));
}

// Render the concert cards
function ConcertGrid(props) {
	if (!props.items || props.items.length == 0) {
		return React.createElement(
			"div",
			{ className: "column is-12" },
			React.createElement("div", { className: "notification is-light" }, "No concerts.")
		);
	}

	return React.createElement(
		React.Fragment,
		null,
		props.items.map(function (concert, index) {
			let id = String(itemID(concert));
			let expired = isExpiredConcert(concert);
			let photo = concertPhoto(concert);
			return React.createElement(
				"div",
				{ className: "column is-half-tablet is-one-third-desktop", key: id || String(index) },
				React.createElement(
					"div",
					{ className: "card ticketmet-concert-card" + (expired ? " ticketmet-concert-card--expired" : "") },
					React.createElement(
						"div",
						{ className: "card-image" },
						React.createElement(
							"figure",
							{ className: "image is-16by9" },
							React.createElement("img", { src: photo, alt: concert.Name || "Concert" })
						)
					),
					React.createElement(
						"div",
						{ className: "card-content" },
						React.createElement("p", { className: "title is-5 mb-2" }, concert.Name),
						React.createElement(
							"p",
							{ className: "subtitle is-6 mb-3" },
							(concert.ArtistName || "Unknown artist") + " · " + (concert.VenueName || "Unknown venue")
						),
						React.createElement("p", { className: "is-size-7 mb-3" }, formatDate(concert.Date)),
						React.createElement("div", { className: "tags" }, renderConcertTags(concert))
					),
					React.createElement(
						"footer",
						{ className: "card-footer" },
						React.createElement(
							"button",
							{
								className: "card-footer-item button is-white",
								type: "button",
								onClick: function () {
									openConcert(id);
								},
							},
							"Open"
						)
					)
				)
			);
		})
	);
}

// Load the next page of concerts
function loadMoreConcerts() {
	loadConcerts(false);
}

// Load more concerts on scroll
function handleConcertScroll() {
	if (concertLoading || !concertHasMore) return;
	let scrollBottom = $(window).scrollTop() + $(window).height();
	let pageBottom = $(document).height() - 200;
	if (scrollBottom >= pageBottom) {
		loadMoreConcerts();
	}
}
