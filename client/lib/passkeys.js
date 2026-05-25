// Render the passkeys list
function PasskeysList(props) {
	if (!props.passkeys || props.passkeys.length == 0) {
		return React.createElement("span", { className: "tag is-light" }, "No passkeys");
	}

	return React.createElement(
		React.Fragment,
		null,
		props.passkeys.map(function (passkey, index) {
			let id = String(passkey.CredentialID || "");
			let signCount = field(passkey, ["SignCount", "signCount"], 0);
			return React.createElement(
				"div",
				{ className: "tag is-info is-light is-medium", key: id || String(index) },
				React.createElement("span", { title: id }, id.substring(0, 18) + "..."),
				React.createElement("span", { className: "ml-2 has-text-grey" }, "#" + signCount),
				React.createElement("button", {
					className: "delete is-small ml-2",
					type: "button",
					onClick: function () {
						deletePasskey(id);
					},
				})
			);
		})
	);
}

// Mount the passkeys list
function renderPasskeysList(passkeys) {
	ReactDOM.render(React.createElement(PasskeysList, { passkeys: passkeys }), document.getElementById("passkeysList"));
}

// Load the current user's passkeys
function loadPasskeys() {
	if (!user) return;

	api("GET", "/api/auth/passkeys")
		.done(function (response) {
			renderPasskeysList(response.Passkeys || []);
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to list passkeys."));
		});
}

// Delete a passkey
function deletePasskey(id) {
	api("DELETE", "/api/auth/passkeys/" + encodeURIComponent(id))
		.done(function () {
			loadPasskeys();
			setMessages("success", "Passkey deleted.");
		})
		.fail(function (xhr) {
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Passkey deletion failed."));
		});
}

// Register a new passkey
function registerPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "This browser does not support passkeys.");
		return;
	}

	setPasskeyLoading("#registerPasskeyButton", true);
	api("POST", "/api/auth/passkeys/register/begin")
		.done(function (options) {
			let publicKey = options.publicKey;
			publicKey.challenge = base64ToBuffer(publicKey.challenge);
			publicKey.user.id = base64ToBuffer(publicKey.user.id);
			if (publicKey.excludeCredentials) {
				for (let i = 0; i < publicKey.excludeCredentials.length; i = i + 1) {
					publicKey.excludeCredentials[i].id = base64ToBuffer(publicKey.excludeCredentials[i].id);
				}
			}

			navigator.credentials.create(options)
				.then(function (credential) {
					return new Promise(function (resolve, reject) {
						api("POST", "/api/auth/passkeys/register/finish", credentialToJSON(credential))
							.done(resolve)
							.fail(function (xhr) {
								if (handleAuthRequired(xhr)) {
									reject(xhr);
									return;
								}
								reject(new Error(errorText(xhr, "server validation refused")));
							});
					});
				})
				.then(function () {
					setPasskeyLoading("#registerPasskeyButton", false);
					loadPasskeys();
					setMessages("success", "Passkey added.");
				})
				.catch(function (error) {
					setPasskeyLoading("#registerPasskeyButton", false);
					if (handleAuthRequired(error)) return;
					setMessages("danger", "Passkey creation failed: " + passkeyError(error));
				});
		})
		.fail(function (xhr) {
			setPasskeyLoading("#registerPasskeyButton", false);
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to start passkey creation."));
		});
}

// Sign in with a passkey
function loginPasskey() {
	if (!window.PublicKeyCredential) {
		setMessages("danger", "This browser does not support passkeys.");
		return;
	}

	setPasskeyLoading("#loginPasskeyButton", true);
	api("POST", "/api/auth/passkeys/login/begin")
		.done(function (options) {
			let publicKey = options.publicKey;
			publicKey.challenge = base64ToBuffer(publicKey.challenge);
			if (publicKey.allowCredentials) {
				for (let i = 0; i < publicKey.allowCredentials.length; i = i + 1) {
					publicKey.allowCredentials[i].id = base64ToBuffer(publicKey.allowCredentials[i].id);
				}
			}

			navigator.credentials.get(options)
				.then(function (credential) {
					return new Promise(function (resolve, reject) {
						api("POST", "/api/auth/passkeys/login/finish", credentialToJSON(credential))
							.done(resolve)
							.fail(function (xhr) {
								if (handleAuthRequired(xhr)) {
									reject(xhr);
									return;
								}
								reject(new Error(errorText(xhr, "server validation refused")));
							});
					});
				})
				.then(function (response) {
					setPasskeyLoading("#loginPasskeyButton", false);
					user = response.User;
					toggleConnected(true);
					setMessages("success", "Signed in with passkey.");
				})
				.catch(function (error) {
					setPasskeyLoading("#loginPasskeyButton", false);
					if (handleAuthRequired(error)) return;
					setMessages("danger", "Passkey sign in failed: " + passkeyError(error));
				});
		})
		.fail(function (xhr) {
			setPasskeyLoading("#loginPasskeyButton", false);
			if (handleAuthRequired(xhr)) return;
			setMessages("danger", errorText(xhr, "Unable to start passkey sign in."));
		});
}

// Toggle a passkey button loading state
function setPasskeyLoading(buttonID, loading) {
	$(buttonID).toggleClass("is-loading", loading);
}

// Convert a credential to JSON
function credentialToJSON(credential) {
	let response = credential.response;
	let json = {
		id: credential.id,
		rawId: bufferToBase64(credential.rawId),
		type: credential.type,
		response: {},
	};

	if (response.clientDataJSON) json.response.clientDataJSON = bufferToBase64(response.clientDataJSON);
	if (response.attestationObject) json.response.attestationObject = bufferToBase64(response.attestationObject);
	if (response.authenticatorData) json.response.authenticatorData = bufferToBase64(response.authenticatorData);
	if (response.signature) json.response.signature = bufferToBase64(response.signature);
	if (response.userHandle) json.response.userHandle = bufferToBase64(response.userHandle);

	return json;
}

// Decode a base64 value into a buffer
function base64ToBuffer(value) {
	let base64 = value.replace(/-/g, "+").replace(/_/g, "/");
	while (base64.length % 4 != 0) {
		base64 = base64 + "=";
	}
	let raw = window.atob(base64);
	let buffer = new ArrayBuffer(raw.length);
	let bytes = new Uint8Array(buffer);
	for (let i = 0; i < raw.length; i = i + 1) {
		bytes[i] = raw.charCodeAt(i);
	}
	return buffer;
}

// Encode a buffer as base64
function bufferToBase64(buffer) {
	let bytes = new Uint8Array(buffer);
	let text = "";
	for (let i = 0; i < bytes.length; i = i + 1) {
		text = text + String.fromCharCode(bytes[i]);
	}
	return window.btoa(text).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

// Read a readable passkey error
function passkeyError(error) {
	if (error && error.message) {
		return error.message;
	}
	return "cancelled or refused by the browser";
}
