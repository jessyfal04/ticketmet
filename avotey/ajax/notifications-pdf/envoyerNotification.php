<?php
	session_start();

	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);
	$typeNotif = strip_tags($_GET['typeNotif']);

	$users = json_decode(file_get_contents('../../data/users.json'), true);

	$status = "";
	$message = "";

	// Check if is logged in
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Not logged in";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);
	$scrutin = $scrutins[$scrutinId];

	// if not found
	if ($scrutin == null) {
		$status = "error";
		$message = "Scrutin not found";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	if ($_SESSION['uuid'] != $scrutin['organizer']) {
		$status = "error";
		$message = "Not authorized";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}


	if (($typeNotif == "ouverture" && $scutin["compteurNotifs"] > 0) || ($typeNotif == "resultats" && $scutin["compteurNotifs"] > 1)) {
		$status = "error";
		$message = "Notifications already sent";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}


	foreach ($scrutin["voters"] as $voterEmail) {
		if ($users[$voterEmail]['notificationToken'] != "") {
			if ($typeNotif == "ouverture") {
				$params = array(
					'token' => $users[$voterEmail]['notificationToken'],
					'title' => "AVotey - Nouveau Scrutin Votable",
					'body' => "Bonjour " . $voterEmail . ", " . $_SESSION['uuid'] . " a créé un nouveau scrutin intitulé " . $scrutin["question"] . " .",
					'image' => "https://jessyfal04.dev/avotey/lib/icon.png",
					'link' => ".."
				);
			}
			else if ($typeNotif == "resultats") {
				$params = array(
					'token' => $users[$voterEmail]['notificationToken'],
					'title' => "AVotey - Résultats du Scrutin",
					'body' => "Bonjour " . $voterEmail . ", " . $_SESSION['uuid'] . " a publié les résultats du scrutin intitulé " . $scrutin["question"] . " .",
					'image' => "https://jessyfal04.dev/avotey/lib/icon.png",
					'link' => ".."
				);
			}
			
			
			$queryString = http_build_query(array('params' => json_encode($params)));
			$requestUrl = "https://jessyfal04.dev/avotey/api/notification/send.php?" . $queryString;
			
			file_get_contents($requestUrl);
		}
	}


	$scrutin["compteurNotifs"] += 1;

	$scrutins[$scrutinId] = $scrutin;
	$writed = file_put_contents('../../data/scrutins.json', json_encode($scrutins, JSON_PRETTY_PRINT));

	if (!$writed) {
		$status = "error";
		$message = "Error writing";
	}
	else {
		$status = "success";
		$message = "Notifications d'ouverture envoyées";
	}

	echo json_encode(array('status' => $status, 'message' => $message));

	