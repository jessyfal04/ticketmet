<?php

	session_start();
	header('Content-Type: application/json');

	$status = "";
	$message = "";
	$data = array();

	// Vérifier si l'utilisateur est connecté
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Pas connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$users = json_decode(file_get_contents('../../data/users.json'), true);
	$user = $users[$_SESSION['uuid']];

	// Vérifier si l'utilisateur existe
	if ($user == null) {
		$status = "error";
		$message = "Utilisateur introuvable";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$status = "success";
	$message = "";
	$data['lists'] = $user['lists'];

	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));