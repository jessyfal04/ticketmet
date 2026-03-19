<?php
	session_start();

	header('Content-Type: application/json');

	$scrutins = json_decode(file_get_contents('../../data/scrutins.json'), true);

	// Check if is logged in
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Not logged in";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$status = "success";
	$message = "";
	$data = array();

	// Check all scrutins the user can vote
	foreach ($scrutins as $scrutinId => $scrutin) {
		$votable = in_array($_SESSION['uuid'], $scrutin['voters']);
		$remaining = $votable ? $scrutin["votesPermis"][$_SESSION['uuid']] - $scrutin["votesUtilises"][$_SESSION['uuid']] : 0;
		$manageable = $_SESSION['uuid'] == $scrutin['organizer'];
		$totalPermis = 0;
		$totalUtilises = 0;

		foreach ($scrutin["votesPermis"] as $value) {
			$totalPermis += $value;
		}
		foreach ($scrutin["votesUtilises"] as $value) {
			$totalUtilises += $value;
		}
		
		if ($votable || $manageable) {
			$scrutin_ = array();
			$scrutin_['id'] = $scrutinId;
			$scrutin_['question'] = $scrutin['question'];
			$scrutin_['organizer'] = $scrutin['organizer'];
			$scrutin_['open'] = $scrutin['open'];

			$scrutin_['votable'] = $votable;
			$scrutin_['manageable'] = $manageable;
			$scrutin_['remaining'] = $remaining;

			$scrutin_["totalPermis"] = $totalPermis;
			$scrutin_["totalUtilises"] = $totalUtilises;

			$data[] = $scrutin_;
		}
	}

	echo json_encode(array('status' => $status, 'message' => $message, 'votables' => $data));