// Import the functions you need from the SDKs you need
import { initializeApp } from "https://www.gstatic.com/firebasejs/10.7.1/firebase-app.js";
import { getMessaging, getToken, deleteToken } from "https://www.gstatic.com/firebasejs/10.7.1/firebase-messaging.js";

const firebaseConfig = {
	apiKey: "REDACTED",
	authDomain: "avotey-c66e4.firebaseapp.com",
	projectId: "avotey-c66e4",
	storageBucket: "avotey-c66e4.appspot.com",
	messagingSenderId: "661537767576",
	appId: "1:661537767576:web:8b1adef4f3c4dd4a7050ed"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const messaging = getMessaging(app);

window.activerNotifications = () => {
	$("#activerNotifs").addClass("is-loading");
	// Request permission and get token.....
	Notification.requestPermission().then((permission) => {
		if (permission === 'granted') {
			console.log('Notification permission granted.');

			navigator.serviceWorker.register("lib/notificationServiceWorker.js").then(registration => {
				getToken(messaging, {
					serviceWorkerRegistration: registration,
					vapidKey: 'REDACTED' }).then((currentToken) => {
					if (currentToken) {
						console.log("Token is: " + currentToken);

						$.ajax({
							method: "GET",
							url: "ajax/notifications-pdf/setNotificationToken.php",
							data: {
								"token": currentToken
							},
							dataType: "json"
						})
						.done(function (e) {
							if (e.status == "success") {
								setMessages("success", e.message);
							} else {
								setMessages("danger", e.message);
							}
							$("#activerNotifs").removeClass("is-loading");

						})
						.fail(function (e) {
							console.log(e);
							$("#activerNotifs").removeClass("is-loading");

						});

						setMessages('success', "Les notifications ont été activées.");
					} else {
						console.log('No registration token available. Request permission to generate one.');
						setMessages('danger', "Aucun token de notification n'est disponible. Veuillez réessayer.");
						$("#activerNotifs").removeClass("is-loading");
					}
				}).catch((err) => {
					console.log('An error occurred while retrieving token. ', err);
					setMessages('danger', "Une erreur est survenue lors de la récupération du token de notification.");
					$("#activerNotifs").removeClass("is-loading");
				});
			});
		} else {
			setMessages('danger', "Impossible d'obtenir la permission de notification.");
			$("#activerNotifs").removeClass("is-loading");
		}
	});

};

window.desactiverNotifications = () => {
	// Unsuscribe from notifications
	navigator.serviceWorker.getRegistrations().then(registrations => {
		for (const registration of registrations) {
			registration.unregister();
		} 
	});

	$.ajax({
		method: "GET",
		url: "ajax/notifications-pdf/setNotificationToken.php",
		data: {
			"token": ""
		},
		dataType: "json"
	})
	.done(function (e) {
		if (e.status == "success") {
			setMessages("success", e.message);
		} else {
			setMessages("danger", e.message);
		}
	})
	.fail(function (e) {
		console.log(e);
	});
}