<?php
	session_start();
	header('Content-Type: application/json');

	$status = "";
	$message = "";

	// Vérifier si l'utilisateur est déjà connecté
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Non connecté";
	} else {
		// Unset the session
		unset($_SESSION['uuid']);
		$status = "success";
		$message = "Déconnexion réussie";
	}

	echo json_encode(array('status' => $status, 'message' => $message));
