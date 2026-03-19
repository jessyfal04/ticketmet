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

	// Vérifier si l'email est valide
	if (!filter_var($getEmail, FILTER_VALIDATE_EMAIL)) {
		$status = "error";
		$message = "Adresse email invalide";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si le mot de passe est valide
	if (strlen($getPassword) < 8) {
		$status = "error";
		$message = "Mot de passe invalide (8 caractères minimum)";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si l'email est déjà utilisé
	foreach ($users as $email => $user) {
		if ($email == $getEmail) {
			$status = "error";
			$message = "Adresse email déjà enregistrée";
			echo json_encode(array('status' => $status, 'message' => $message));
			exit();
		}
	}

	// Ajouter l'utilisateur au tableau des utilisateurs
	$users[$getEmail] = array(
		'password' => password_hash($getPassword, PASSWORD_DEFAULT),
		'lists' => array(),
		'notificationToken' => ""
	);

	// Écrire le tableau des utilisateurs dans le fichier
	$writed = file_put_contents('../../data/users.json', json_encode($users, JSON_PRETTY_PRINT));

	if (!$writed) {
		$status = "error";
		$message = "Erreur lors de l'écriture du fichier";
	} else {
		$_SESSION['uuid'] = $getEmail;
		$data['uuid'] = $getEmail;
		$status = "success";
		$message = "Compte créé avec succès";
	}

	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));