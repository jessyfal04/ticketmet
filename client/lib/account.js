// Account
// Check the backend health endpoint
function checkHealth() {
	$.ajax({ method: "GET", url: "/healthz" })
		.done(function () {
			$("#healthStatus").removeClass("is-warning is-danger").addClass("is-success").text("");
		})
		.fail(function () {
			$("#healthStatus").removeClass("is-warning is-success").addClass("is-danger").text("");
		});
}

// Check the current session
function checkSession() {
	api("GET", "/api/auth/me")
		.done(function (response) {
			user = response.User;
			toggleConnected(true);
		})
		.fail(function () {
			user = null;
			toggleConnected(false);
		});
}

// Sign in with email and password
function login() {
	let email = $("#email").val();
	let password = $("#password").val();

	api("POST", "/api/auth/login", { email: email, password: password })
		.done(function (response) {
			user = response.User;
			toggleConnected(true);
			setMessages("success", "Signed in.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Sign in failed."));
		});
}

// Create a new account
function register() {
	let email = $("#email").val();
	let password = $("#password").val();

	api("POST", "/api/auth/register", { email: email, password: password })
		.done(function (response) {
			user = response.User;
			toggleConnected(true);
			setMessages("success", "Account created.");
		})
		.fail(function (xhr) {
			setMessages("danger", errorText(xhr, "Registration failed."));
		});
}

// Sign out the current user
function logout() {
	$.ajax({ method: "POST", url: "/api/auth/logout" })
		.always(function () {
			user = null;
			toggleConnected(false);
			showView("accountView");
			setMessages("success", "Signed out.");
		});
}

// Delete the current account
function unregisterAccount() {
	if (!user) return;
	if (!window.confirm("Delete this account and all its data?")) return;

	api("DELETE", "/api/auth/unregister", { password: $("#unregisterPassword").val() })
		.done(function () {
			user = null;
			toggleConnected(false);
			showView("accountView");
			setMessages("success", "Account deleted.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Account deletion failed."));
		});
}

// Verify whether the email exists
function emailExists() {
	let email = $("#email").val();
	if (!email) return;

	api("GET", "/api/auth/email-exists", { email: email })
		.done(function (response) {
			if (response.Exists) {
				$("#emailHelp").text("Email already registered.");
				$("#registerButton").prop("disabled", true);
				$("#loginButton").prop("disabled", false);
			} else {
				$("#emailHelp").text("Email available.");
				$("#registerButton").prop("disabled", false);
				$("#loginButton").prop("disabled", true);
			}
		});
}

// Toggle the connected state in the UI
function toggleConnected(connected) {
	$(".if-login").css("display", connected ? "" : "none");
	$(".if-not-login").css("display", connected ? "none" : "");
	$("#connectedEmail").text(user ? user.Email : "");
	if (!connected) {
		profileData = null;
	}
	renderAccount();
	renderAlertBlock();
	renderConcertActions();
	if (connected) {
		loadPasskeys();
		loadProfile();
	}
}

// Render the account section
function renderAccount() {
	if (!user) {
		$("#accountID").text("");
		$("#accountEmail").text("");
		$("#profileSNS").val("");
		$("#profileSNS").removeClass("is-success");
		$("#profileSNS").removeClass("is-danger");
		$("#passkeysList").html("");
		$("#profileFavorites").html("");
		$("#profileWT").html("");
		$("#profileAlerts").html("");
		$("#unregisterPassword").val("");
		return;
	}

	$("#accountID").text(user.ID);
	$("#accountEmail").text(user.Email);
	$("#unregisterPassword").val("");
	renderProfileData();
}

// Load the user profile
function loadProfile() {
	if (!user) return;

	api("GET", "/api/me")
		.done(function (response) {
			profileData = response;
			renderProfileData();
			renderAlertBlock();
			renderConcertActions();
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to load profile."));
		});
}

// Save the SNS profile fields
function saveSNS() {
	if (!user) return;
	window.clearTimeout(saveSNSDelay);

	let sns = $("#profileSNS").val().split("\n").map(function (value) {
		return value.trim();
	}).filter(function (value) {
		return value != "";
	});

	api("PATCH", "/api/me", { SNS: sns })
		.done(function (response) {
			profileData = response;
			renderProfileData();
			loadConcertFeatures();
			$("#profileSNS").removeClass("is-danger");
			$("#profileSNS").addClass("is-success");
			setMessages("success", "SNS saved.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			$("#profileSNS").removeClass("is-success");
			$("#profileSNS").addClass("is-danger");
			setMessages("danger", errorText(xhr, "Unable to save SNS."));
		});
}

// Read the SNS form values
function profileSNSInput() {
	$("#profileSNS").removeClass("is-success");
	$("#profileSNS").addClass("is-danger");
	scheduleSaveSNS();
}

// Schedule the SNS save
function scheduleSaveSNS() {
	if (!user) return;
	window.clearTimeout(saveSNSDelay);
	saveSNSDelay = window.setTimeout(saveSNS, 500);
}

// Render the profile summary
function renderProfileData() {
	if (!profileData) {
		$("#profileFavorites").html(`<span class="has-text-grey">No data loaded.</span>`);
		$("#profileWT").html(`<span class="has-text-grey">No data loaded.</span>`);
		$("#profileAlerts").html(`<span class="has-text-grey">No data loaded.</span>`);
		return;
	}

	let sns = profileData.SNS || [];
	$("#profileSNS").val(sns.join("\n"));
	$("#profileSNS").removeClass("is-danger");
	$("#profileSNS").addClass("is-success");

	renderProfileConcertList("#profileFavorites", profileData.Favorites || [], "No favorites yet.");
	renderProfileWTList(profileData.WT || []);
	renderProfileAlerts(profileData.Alerts || []);
}

// Render the alert controls for the current filters
function AlertBlock(props) {
	if (!props.user) {
		return React.createElement("span", { className: "has-text-grey" }, "Sign in to create artist or venue alerts.");
	}

	let artistButtonClass = "button " + (props.artistAlert ? "is-primary is-light" : "is-primary");
	let venueButtonClass = "button " + (props.venueAlert ? "is-primary is-light" : "is-primary");

	return React.createElement(
		"div",
		null,
		React.createElement(
			"div",
			{ className: "content mb-3" },
			React.createElement("p", { className: "mb-1" }, React.createElement("strong", null, "Artist:"), " ", props.artistName),
			React.createElement("p", null, React.createElement("strong", null, "Venue:"), " ", props.venueName)
		),
		React.createElement(
			"div",
			{ className: "buttons" },
			React.createElement(
				"button",
				{
					className: artistButtonClass,
					type: "button",
					disabled: !props.artistSelected,
					onClick: function () {
						createAlertFromSelection("artist");
					},
				},
				props.artistAlert ? "Remove artist alert" : "Alert this artist"
			),
			React.createElement(
				"button",
				{
					className: venueButtonClass,
					type: "button",
					disabled: !props.venueSelected,
					onClick: function () {
						createAlertFromSelection("venue");
					},
				},
				props.venueAlert ? "Remove venue alert" : "Alert this venue"
			)
		)
	);
}

// Mount the alert controls
function renderAlertBlock() {
	let target = document.getElementById("alertBlock");
	if (!target || !window.React || !window.ReactDOM) return;

	let artistName = selectedArtistID ? selectedName(artists, selectedArtistID) : "All artists";
	let venueName = selectedVenueID ? selectedName(venues, selectedVenueID) : "All venues";
	let artistAlert = getAlert("artist", selectedArtistID);
	let venueAlert = getAlert("venue", selectedVenueID);
	ReactDOM.render(
		React.createElement(AlertBlock, {
			user: user,
			artistName: artistName,
			venueName: venueName,
			artistAlert: artistAlert,
			venueAlert: venueAlert,
			artistSelected: !!selectedArtistID,
			venueSelected: !!selectedVenueID,
		}),
		target
	);
}

// Read the selected item name from <select>
function selectedName(items, id) {
	for (let i = 0; i < items.length; i = i + 1) {
		if (String(itemID(items[i])) == String(id)) {
			return itemName(items[i]);
		}
	}
	return "Selected #" + id;
}

// Render the favorite concert list
function ProfileConcertList(props) {
	if (!props.items || props.items.length == 0) {
		return React.createElement("span", { className: "has-text-grey" }, props.emptyText);
	}

	return React.createElement(
		React.Fragment,
		null,
		props.items.map(function (concert, index) {
			return React.createElement(
				"div",
				{ className: "tags has-addons mb-1", key: String(itemID(concert) || index) },
				React.createElement("span", { className: "tag is-primary is-light" }, "Favorite"),
				React.createElement(
					"button",
					{
						className: "tag is-light",
						type: "button",
						onClick: function () {
							openConcert(itemID(concert));
						},
					},
					concert.Name
				)
			);
		})
	);
}

// Mount the favorite concert list
function renderProfileConcertList(target, items, emptyText) {
	let node = document.getElementById(String(target).replace(/^#/, ""));
	ReactDOM.render(React.createElement(ProfileConcertList, { items: items, emptyText: emptyText }), node);
}

// Render the WTB/WTS list
function ProfileWTList(props) {
	if (!props.items || props.items.length == 0) {
		return React.createElement("span", { className: "has-text-grey" }, "No WTB/WTS yet.");
	}

	return React.createElement(
		React.Fragment,
		null,
		props.items.map(function (item, index) {
			let concert = item.Concert || {};
			return React.createElement(
				"div",
				{ className: "tags has-addons mb-1", key: String(itemID(concert) || index) },
				React.createElement("span", { className: "tag is-primary is-light" }, String(item.Type).toUpperCase()),
				React.createElement(
					"button",
					{
						className: "tag is-light",
						type: "button",
						onClick: function () {
							openConcert(itemID(concert));
						},
					},
					concert.Name
				)
			);
		})
	);
}

// Mount the WTB/WTS list
function renderProfileWTList(items) {
	ReactDOM.render(React.createElement(ProfileWTList, { items: items }), document.getElementById("profileWT"));
}

// Render the alert list
function ProfileAlerts(props) {
	if (!props.items || props.items.length == 0) {
		return React.createElement("span", { className: "has-text-grey" }, "No alerts yet.");
	}

	return React.createElement(
		React.Fragment,
		null,
		props.items.map(function (alert, index) {
			let type = String(alert.TargetType || "target");
			type = type.charAt(0).toUpperCase() + type.slice(1);
			return React.createElement(
				"div",
				{ className: "tags has-addons mb-1", key: String(alert.ID || index) },
				React.createElement("span", { className: "tag is-info is-light" }, type),
				React.createElement("span", { className: "tag" }, alert.TargetName || "target"),
				React.createElement("a", {
					className: "tag is-delete",
					href: "#",
					onClick: function (event) {
						event.preventDefault();
						deleteAlert(alert.ID);
					},
				})
			);
		})
	);
}

// Mount the alert list
function renderProfileAlerts(items) {
	ReactDOM.render(React.createElement(ProfileAlerts, { items: items }), document.getElementById("profileAlerts"));
}

// Create or remove an alert from the current selection
function createAlertFromSelection(targetType) {
	let targetID = targetType == "artist" ? selectedArtistID : selectedVenueID;
	if (!targetID) {
		setMessages("warning", "Select a " + targetType + " first.");
		return;
	}
	toggleAlert(targetType, targetID);
}

// Create or remove an alert
function createAlert(targetType, targetID) {
	if (!user) {
		setMessages("warning", "Sign in first.");
		return;
	}

	api("POST", "/api/alerts?targetType=" + encodeURIComponent(targetType) + "&targetId=" + encodeURIComponent(targetID))
		.done(function () {
			loadProfile();
			setMessages("success", "Alert created.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to create alert."));
		});
}

// Toggle an alert on the API
function toggleAlert(targetType, targetID) {
	let alert = getAlert(targetType, targetID);
	if (alert) {
		deleteAlert(alert.ID);
		return;
	}
	createAlert(targetType, targetID);
}

// Delete an alert
function deleteAlert(id) {
	api("DELETE", "/api/alerts/" + encodeURIComponent(id))
		.done(function () {
			loadProfile();
			setMessages("success", "Alert deleted.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to delete alert."));
		});
}

// Find the matching alert
function getAlert(targetType, targetID) {
	if (!profileData || !profileData.Alerts) return null;
	for (let i = 0; i < profileData.Alerts.length; i = i + 1) {
		let alert = profileData.Alerts[i];
		if (String(alert.TargetType) == String(targetType) && String(alert.TargetID) == String(targetID)) {
			return alert;
		}
	}
	return null;
}
