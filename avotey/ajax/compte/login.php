<?php
	session_start();
	header('Content-Type: application/json');

	$getEmail = strip_tags($_GET['email']);
	$getPassword = strip_tags($_GET['password']);

	$users = json_decode(file_get_contents('../../data/users.json'), true);

	$status = "";
	$message = "";
	$data = array();

	// Vérifier si l'utilisateur est déjà connecté
	if (isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Déjà connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si l'email et le mot de passe correspondent
	$match = false;
	foreach ($users as $email => $user) {
		// Si l'email et le mot de passe correspondent
		if ($email == $getEmail && password_verify($getPassword, $user['password'])) {
			$match = true;
			$_SESSION['uuid'] = $getEmail;
			$data['uuid'] = $getEmail;

			$status = "success";
			$message = "Connexion réussie";

			echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));
			exit();
		}
	}
	

	// Retourner le statut et le message et les datas
	$status = "error";
	$message = "Email ou mot de passe incorrect";
	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));