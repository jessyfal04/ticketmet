<?php
	session_start();
	header('Content-Type: application/json');

	$getEmail = strip_tags($_GET['email']);

	$users = json_decode(file_get_contents('../../data/users.json'), true);

	$status = "success";
	$message = "Adresse email non enregistrée";
	$data = 0;

	// On vérifie si l'adresse email existe déjà
	foreach ($users as $email => $user) {
		if ($email == $getEmail) {
			$status = "success";
			$message = "Adresse email déjà enregistrée";
			$data = 1;
		}
	}

	// Retourner le statut et le message et les datas
	echo json_encode(array(
		"status" => $status,
		"message" => $message,
		"data" => $data
	));