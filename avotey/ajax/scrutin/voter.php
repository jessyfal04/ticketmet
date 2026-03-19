<?php
	session_start();

	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);
	$choice = strip_tags($_GET['choice']);

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	$status = "";
	$message = "";

	// Check if is logged in
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Pas connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Check if the choice is valid
	if (strlen($choice) < 1) {
		$status = "error";
		$message = "Choix invalide";
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

	
	if (!in_array($_SESSION['uuid'], $scrutin['voters'])) {
		$status = "error";
		$message = "Vous n'êtes pas autorisé à voter pour ce scrutin";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}
	

	if ($scrutin['open'] == false) {
		$status = "error";
		$message = "Scrutin déjà fermé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Check if vote remaining
	if($scrutin["votesPermis"][$_SESSION['uuid']] - $scrutin["votesUtilises"][$_SESSION['uuid']] == 0) {
		$status = "error";
		$message = "Vous avez déjà voté le nombre de fois autorisé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	} else {
		$scrutin["votesUtilises"][$_SESSION['uuid']] += 1;
	}

	$status = "success";
	$message = "";

	$scrutin['votes'][] = $choice;

	$scrutins[$scrutinId] = $scrutin;
	$writed = file_put_contents('../../data/scrutins.json', json_encode($scrutins, JSON_PRETTY_PRINT));

	if (!$writed) {
		$status = "error";
		$message = "Erreur lors de l'écriture du fichier de scrutin";
	} else {
		$status = "success";
		$message = "Vote enregistré";
	}

	echo json_encode(array('status' => $status, 'message' => $message));
