<?php
	session_start();

	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);
	$resultats = json_decode($_GET['resultats']);

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

	$scrutin = $scrutins[$scrutinId];

	// if not found
	if ($scrutin == null) {
		$status = "error";
		$message = "Scrutin introuvable";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	if ($_SESSION['uuid'] != $scrutin['organizer']) {
		$status = "error";
		$message = "Pas organisateur";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	if ($scrutin['open'] == false) {
		$status = "error";
		$message = "Scrutin déjà fermé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$scrutin['open'] = false;
	$scrutin['resultats'] = $resultats;
	$scrutins[$scrutinId] = $scrutin;

	$status = "success";
	$message = "Resultats publiés";

	$writed = file_put_contents('../../data/scrutins.json', json_encode($scrutins, JSON_PRETTY_PRINT));

	if ($writed === false) {
		$status = "error";
		$message = "Erreur lors de l'écriture du fichier de scrutin";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}
	else {
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}
	

