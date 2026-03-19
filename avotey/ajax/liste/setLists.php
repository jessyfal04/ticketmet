<?php

	session_start();
	header('Content-Type: application/json');

	$lists = json_decode($_GET['lists']);

	$status = "";
	$message = "";

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

	// Vérifier si les listes sont bien formatées
	foreach ($lists as $titreList => $list) {
		// if titreList is empty
		if (empty($titreList)) {
			$status = "error";
			$message = "Titre vide";
			echo json_encode(array('status' => $status, 'message' => $message));
			exit();
		}

		// for each votant
		foreach ($list as $email => $permis) {
			// if email is empty
			if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
				$status = "error";
				$message = "Email invalide";
				echo json_encode(array('status' => $status, 'message' => $message));
				exit();
			}

			// if permis inferieur à 0 or superieur à 2
			if ($permis > 3 || $permis < 1) {
				$status = "error";
				$message = "Procurations invalides (doit être entre 0 et 2)";
				echo json_encode(array('status' => $status, 'message' => $message));
				exit();
			}
		}
	}
	

	$users[$_SESSION['uuid']]['lists'] = $lists;

	$writed = file_put_contents('../../data/users.json', json_encode($users, JSON_PRETTY_PRINT));

	if ($writed) {
		$status = "success";
		$message = "Liste mise à jour avec succès";
	} else {
		$status = "error";
		$message = "Erreur lors de l'écriture du fichier de liste";
	}


	echo json_encode(array('status' => $status, 'message' => $message));