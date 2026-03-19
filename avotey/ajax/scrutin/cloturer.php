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

	$data["choices"] = $scrutin["choices"];
	$data["votes"] =  $scrutin["votes"];
	$data["systemeVote"] = $scrutin["systemeVote"];
	
	// On mélangera les votes pour éviter de voir qui a voté quoi
	shuffle($data["votes"]);

	$status = "success";
	$message = "Choix et votes récupérés";

	echo json_encode(array('status' => $status, 'message' => $message, 'data' => $data));
	exit();
