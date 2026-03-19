<?php

	session_start();

	header('Content-Type: application/json');

	$getToken = strip_tags($_GET['token']);

	$status = "";
	$message = "";

	// Check if is logged in
	if (!isset($_SESSION['uuid'])) {
		$status = "error";
		$message = "Not logged in";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}

	$users = json_decode(file_get_contents('../../data/users.json'), true);
	$user = $users[$_SESSION['uuid']];

	// if not found
	if ($user == null) {
		$status = "error";
		$message = "User not found";
		echo json_encode(array('status' => $status, 'message' => $message));
		exit();
	}
	

	// If remove token, get the token and remove for all users who have it
	$removeToken = false;
	if ($getToken == "") {
		$removeToken = true;
		$originalToken = $user['notificationToken'];

		foreach ($users as $uuid => $user) {
			if ($user['notificationToken'] == $originalToken) {
				$users[$uuid]['notificationToken'] = "";
			}
		}
	}

	$users[$_SESSION['uuid']]['notificationToken'] = $getToken;


	$writed = file_put_contents('../../data/users.json', json_encode($users, JSON_PRETTY_PRINT));

	if ($writed) {
		$status = "success";
		$message = $removeToken ? "Notifications désactivées pour cet appareil" : "Notifications activées pour ce compte";
	} else {
		$status = "error";
		$message = "Error writing file";
	}


	echo json_encode(array('status' => $status, 'message' => $message));