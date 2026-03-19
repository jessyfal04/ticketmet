<?php
	session_start();

	header('Content-Type: application/json');

	$scrutinId = strip_tags($_GET['scrutinId']);

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	$status = "";
	$message = "";
	$data = array();

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

	if (!in_array($_SESSION['uuid'], $scrutin['voters'])){
		$status = "error";
		$message = "Pas votant";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	if ($scrutin['open'] == false) {
		$status = "error";
		$message = "Scrutin déjà fermé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	if($scrutin["votesPermis"][$_SESSION['uuid']] - $scrutin["votesUtilises"][$_SESSION['uuid']] == 0) {
		$status = "error";
		$message = "Vous avez déjà voté le nombre de fois autorisé";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$status = "success";
	$message = "";

	$data['question'] = $scrutin['question'];
	$data['choices'] = $scrutin['choices'];
	$data['publicKey'] = $scrutin['publicKey'];
	$data['systemeVote'] = $scrutin['systemeVote'];
	$data['who'] = "Vous voterez en tant que " . $_SESSION['uuid'] . " (" . $scrutin["votesPermis"][$_SESSION['uuid']] - $scrutin["votesUtilises"][$_SESSION['uuid']] . " votes restants).";

	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));