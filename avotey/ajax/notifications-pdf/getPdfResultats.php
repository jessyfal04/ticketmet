<?php
	session_start();
	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);

	$users = json_decode(file_get_contents('../../data/users.json'), true);

	$status = "";
	$message = "";

	// Check if is logged in
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Pas connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	$scrutin = $scrutins[$scrutinId];

	// if not found
	if ($scrutin == null) {
		$status = "error";
		$message = "Scrutin introuvable";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$votable = false;
	foreach ($scrutin["voters"] as $voterEmail) {
		if ($voterEmail == $_SESSION['uuid']) {
			$votable = true;
			break;
		}
	}

	if (!$votable && $_SESSION['uuid'] != $scrutin['organizer']) {
		$status = "error";
		$message = "Ni votant ni organisateur";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Check if scrutin is closed
	if ($scrutin["open"] != false) {
		$status = "error";
		$message = "Scrutin pas fermé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}
	

	// Create new PDF document via filegetcontents
	$params = array(
		'author' => $_SESSION['uuid'] . " (AVotey)",
		'title' => "Résultats du scrutin " . $scrutinId,
		'data' => json_encode(Array (
			$scrutinId => $scrutin
		)),
		'type' => "avoteyResultats"
	);
	// without params, each key is a GET parameter
	$queryString = http_build_query($params);
	$requestUrl = "https://jessyfal04.dev/avotey/api/jsonToPdf/generatePdf.php?" . $queryString;

	$readed = file_get_contents($requestUrl);

	// Check if PDF is created
	if ($readed == null) {
		$status = "error";
		$message = "PDF non créé";
		echo json_encode(array('status' => $status, 'message' => $message, 'data' => $readed));
		exit();
	}

	// Send PDF to user

	$data = Array(
		"pdf" => base64_encode($readed)
	);

	$status = "success";
	$message = "PDF créé";
	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));
