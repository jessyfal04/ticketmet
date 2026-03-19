<?php
	session_start();

	header('Content-Type: application/json');

	$question = strip_tags($_GET['question']);
	$choices = json_decode($_GET['choices']);
	$voters = json_decode($_GET['voters']);
	$publicKey = strip_tags($_GET['publicKey']);
	$systemeVote = strip_tags($_GET['systemeVote']);

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	$status = "";
	$message = "";
	$data = array();

	// Vérifier si l'utilisateur est connecté
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Non connecté";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si la question est valide
	if (strlen($question) < 1) {
		$status = "error";
		$message = "Question invalide (1 caractères minimum)";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si les choix sont au moins 2
	if (count($choices) < 2) {
		$status = "error";
		$message = "Il doit y avoir au moins 2 choix";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si les choix sont au moins 1 caractère
	foreach ($choices as $choice) {
		if (strlen($choice) < 1) {
			$status = "error";
			$message = "Les choix doivent faire au moins 1 caractère";
			echo json_encode(array('status' => $status, 'message' => $message));
			exit();
		}
	}

	// Vérifier si les choix sont uniques
	if (count($choices) != count(array_unique($choices))) {
		$status = "error";
		$message = "Les choix doivent être uniques";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si les votants sont au moins 1
	if (count($voters) < 1) {
		$status = "error";
		$message = "Il doit y avoir au moins 1 votant";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Vérifier si la clé publique est valide
	if (strlen($publicKey) < 1) {
		$status = "error";
		$message = "Public key incorrect";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	// Créer un tableau pour les votants
	$votersEmail = [];
	$votesPermis = array();
	$votesUtilises = array();
	
	foreach ($voters as $voter) {
		$voter = (array) $voter;

		// Vérifier si les votants ont une adresse email valide
		if (!filter_var($voter["email"], FILTER_VALIDATE_EMAIL)) {
			$status = "error";
			$message = "Adresse email invalide";
			echo json_encode(array('status' => $status, 'message' => $message));
			exit();
		}

		$votersEmail[] = $voter["email"];
		$votesUtilises[$voter["email"]] = 0;
		$votesPermis[$voter["email"]] = intval($voter["procuration"]) + 1;

		// Vérifier si le nombre de procuration est valide
		if ($votesPermis[$voter["email"]] > 3 || $votesPermis[$voter["email"]] < 1) {
			$status = "error";
			$message = "Nombre de procuration invalide (0 à 2)";
			echo json_encode(array('status' => $status, 'message' => $message));
			exit();
		}

	}

	// Vérifier si les votants sont uniques
	if (count($votersEmail) != count(array_unique($votersEmail))) {
		$status = "error";
		$message = "Les votants doivent être uniques";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	
	$scrutin = array(
		"organizer" => $_SESSION['uuid'],
		"question" => $question,
		"choices" => $choices,
		"systemeVote" => $systemeVote,

		"voters" => $votersEmail,
		"votesPermis" => $votesPermis,
		"votesUtilises" => $votesUtilises,

		"compteurNotifs" => 0,
		"open" => true,
		"votes" => [],

		"publicKey" => $publicKey,
		"resultats" => []
	);

	// Générer un id unique pour le scrutin et l'ajouter au tableau des scrutins
	$scrutinId = uniqid("S-");
	$scrutins[$scrutinId] = $scrutin;

	// Écrire le tableau des scrutins dans le fichier scrutins.json
	$writed = file_put_contents('../../data/scrutins.json', json_encode($scrutins, JSON_PRETTY_PRINT));

	if (!$writed) {
		$status = "error";
		$message = "Erreur lors de l'édition du fichier scrutins.json";
	}

	else {
		$status = "success";
		$message = "Scrutin créé avec succès";
		$data["scrutinId"] = $scrutinId;
	}

	// Retourner le statut, le message et les données
	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));

