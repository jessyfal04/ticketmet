<?php
	session_start();

	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	$status = "";
	$message = "";

	// Vérifier si l'utilisateur est connecté
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Pas connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$scrutin = $scrutins[$scrutinId];

	// Si le scrutin n'est pas trouvé
	if ($scrutin == null) {
		$status = "error";
		$message = "Scrutin non trouvé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	if ($_SESSION['uuid'] != $scrutin['organizer']) {
		$status = "error";
		$message = "Pas organisateur";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	
	// Retirer le scrutin
	unset($scrutins[$scrutinId]);

	$writed = file_put_contents('../../data/scrutins.json', json_encode($scrutins, JSON_PRETTY_PRINT));

	if ($writed === false) {
		$status = "error";
		$message = "Erreur lors de l'écriture du fichier de scrutin";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	} else {
		$status = "success";
		$message = "Scrutin détruit avec succès";

		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}
