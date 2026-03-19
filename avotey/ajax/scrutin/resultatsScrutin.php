<?php

	session_start();

	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	$status = "";
	$message = "";
	$data = array();

	// Check if is logged in
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Pas connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$scrutin = $scrutins[$scrutinId];

	// if not found
	if ($scrutin == null) {
		$status = "error";
		$message = "Scrutin introuvable";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Check if scrutin closed
	if ($scrutin['open'] == true) {
		$status = "error";
		$message = "Scrutin pas encore fermé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$votable = in_array($_SESSION['uuid'], $scrutin['voters']);
	$manageable = $_SESSION['uuid'] == $scrutin['organizer'];

	$totalPermis = 0;
	$totalUtilises = 0;

	foreach ($scrutin["votesPermis"] as $value) {
		$totalPermis += $value;
	}
	foreach ($scrutin["votesUtilises"] as $value) {
		$totalUtilises += $value;
	}

	// Check if the user can access results
	if (!$votable && !$manageable) {
		$status = "error";
		$message = "Ni votant ni organisateur";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$status = "success";
	$message = "";

	$data['question'] = $scrutin['question'];
	$data['organizer'] = $scrutin['organizer'];

	$data['resultats'] = $scrutin['resultats'];
	$data['systemeVote'] = $scrutin['systemeVote'];
	
	$data['totalPermis'] = $totalPermis;
	$data['totalUtilises'] = $totalUtilises;

	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));