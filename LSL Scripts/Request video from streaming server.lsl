// Streaming server controller controller
// On touch, shows a counter, counting downwards to zero, and then reads
// the notecard in inventory.
// Code based on original development by Cyberstorm Martinez for Beta Technologies.
// © 2023 by Gwyneth Llewelyn for Beta Technologies. All rights reserved.
//
// New tech using "StreamDude" to request a specific file to be streamed. (gwyneth 20230811)

string video		= "Intro-720p-h265.mp4";
string path			= "/var/www/clients/client6/web14/home/betafiles/data/beta-technologies/Universidade de Aveiro/LOCUS Project in Amiais/Panels SL/Painel_Intro/";
string streamerAPI  = "https://streaming.betatechnologies.info/StreamDude/api";
string streamerURL	= "rtsp://video.betatechnologies.info:5544/"; // + video
string lalMasterKey = "WH13TR5QRC4TFH06";

/*
 *  StreamDude works as a two-step process: first, we send the object ID and get an authentication
 *  token, which can then be used to request music and videos to be streamed.
 */
string PIN = "8765";	// 4-digit PIN code, can also be used for some hashes, so long as it's unique
string token;			// token received after authentication
key reqAuth;			// authorisation request, sending our PIN, and receiving a token
key reqPlay;			// request to stream a video
key reqDelete;			// request to delete token

string gName;   		// for the notecard reading
integer gLine;  		// current notecard line (starts at zero)
key gQueryID;   		// handler for the database event retrieving notecard
key avatarKey;			// avatar UUID who touched the object
string avatarName;		// avatar name who  "  "	"   "  "
integer seconds;		// downwards counter
integer BT_DEBUG_CHANNEL = -534133952;	// moving out of the real DEBUG_CHANNEL to avoid floating error icons


default
{
	state_entry()
	{
		llSetText(llGetObjectDesc(), <0.8, 0.6, 0.0>, 1.0);
		llSetClickAction(CLICK_ACTION_TOUCH);
		llSetTouchText("▶︎ Video");
		llParcelMediaCommandList([PARCEL_MEDIA_COMMAND_STOP]);
		avatarKey = NULL_KEY;
	}

	touch_start(integer total_number)
	{
		// we will only deal with one user at the time
		avatarKey = llDetectedKey(0);
		avatarName = llDetectedName(0);
		llParcelMediaCommandList([PARCEL_MEDIA_COMMAND_PLAY]);
		llSay(PUBLIC_CHANNEL, avatarName + ", " + llGetObjectDesc());
		state connectToStreamer;
	}

	changed(integer c)
	{
		if (c & CHANGED_INVENTORY)
		{
			llResetScript();
		}
	}
}

// deal with the protocol of connecting with the streaming server
state connectToStreamer
{
	state_entry()
	{
		llSetClickAction(CLICK_ACTION_DISABLED);
		llSetTouchText("Wait!");

		// this may fail, or no videos get streamed, or something;
		// so do a full reset after 15 minutes
		llSetTimerEvent(900.0);

		// exchange PIN for an auth token:
		string request = "objectPIN=" + PIN + "&masterKey=" + lalMasterKey
			+ "&avatarName=" + avatarName
			+ "&avatarKey=" + (string)avatarKey;
		reqAuth = llHTTPRequest(streamerAPI + "/auth", [
				HTTP_METHOD, "POST",
				HTTP_MIMETYPE, "application/x-www-form-urlencoded",
				HTTP_ACCEPT, "text/plain", // avoid HTML, since we might not have enough memory to display that
				HTTP_VERBOSE_THROTTLE, FALSE
			],
			request);
		if (reqAuth == NULL_KEY)
		{
			// warn user that we're throttling requests
			llRegionSayTo(avatarKey, PUBLIC_CHANNEL, "Too many simultaneous requests, please try in a moment again.");
			state default;
		}
	}

	http_response(key request_id, integer status, list metadata, string body)
	{
		if (request_id == NULL_KEY)
		{
			llRegionSay(BT_DEBUG_CHANNEL, "Weird, a null HTTP request received just now...");
		}
		else if (request_id == reqAuth)
		{
			if (status == 200)
			{
				// We got the expected response; save token
				// TODO: handle simultaneous requests for more than one token
				// Right now, we restrict this to only one avatar touching (but all can view in sync)
				token = body;
				llRegionSay(BT_DEBUG_CHANNEL, "Token received for avatar '" + avatarName + "': "  + token);
				llRegionSay(BT_DEBUG_CHANNEL, "Requesting '" + path + video + "' for streaming.");
				// Now make the request for the video, using this token:
				string request = "objectPIN=" + PIN + "&masterKey=" + lalMasterKey + "&token=" + token
					+ "&avatarName="+ avatarName + "&avatarKey=" + (string)avatarKey
					+ "&filename=" + path + video;
				reqPlay = llHTTPRequest(streamerAPI + "/play", [
						HTTP_METHOD, "POST",
						HTTP_MIMETYPE, "application/x-www-form-urlencoded",
						HTTP_ACCEPT, "text/plain", // avoid HTML, since we might not have enough memory to display that
						HTTP_VERBOSE_THROTTLE, FALSE
					],
					request);
			}
			else
			{
				llRegionSay(BT_DEBUG_CHANNEL, "Error requesting token for avatar '" + avatarName + "': "
					+ (string)status + ": " + body);
				llRegionSayTo(avatarKey, PUBLIC_CHANNEL, "Sorry, our streaming server seems to be down. Please try later!");
				token = "";	// better to reset it
			}
		}
		// request sent from StreamDude
		else if (request_id == reqPlay)
		{
			if (status == 200)
			{
				llRegionSay(BT_DEBUG_CHANNEL, "Streaming request accepted for:" + path + video);
				// Set parcel streaming URL:
				llParcelMediaCommandList([PARCEL_MEDIA_COMMAND_URL, streamerURL + video,
					PARCEL_MEDIA_COMMAND_TYPE, "video/quicktime",	// possibly not necessary
					PARCEL_MEDIA_COMMAND_TIME, 0.0,
					PARCEL_MEDIA_COMMAND_PLAY
				]);
				// now wait for the streamer to do its streaming magic:
				state wait_for_video;
			}
			else
			{
				llRegionSay(BT_DEBUG_CHANNEL, "Fail! Streaming request REJECTED for: '" + path + video
					+ "', error - " + (string)status + ": " + body);
			}
		}
	}

	timer()
	{
		llRegionSay(BT_DEBUG_CHANNEL, "Communications with the outside world seem to be broken; doing a full reset now.");
		llSay(PUBLIC_CHANNEL, "Video streamer seems to be down; resetting...");
		llResetScript();
	}
}

state wait_for_video
{
	state_entry()
	{
		llSetClickAction(CLICK_ACTION_PLAY);
		llSetTouchText("▶︎/❚❚");
		/* seconds = 60;	// video is actually 64 seconds long */
		seconds = 90;   	// so we give it a bit of extra margin...
		llSetTimerEvent(1.0);
	}

	timer()
	{
		if (seconds <= 0) {
			llSetTimerEvent(0.0);
			state sayNotecard;
		}

		llSetText("Time until hint is revealed\n☞ " + (string)seconds +"s ☜", <0.6,0.8,0.0>, 1.0);
		seconds--;
	}

	changed(integer c)
	{
		if (c & CHANGED_INVENTORY)
		{
			llResetScript();
		}
	}
}

state sayNotecard
{
	state_entry()
	{
		llSetText("Hint sent to " + avatarName, <0.8,0.6,0.0>, 1.0);
		if (avatarKey == NULL_KEY) {
			llWhisper(PUBLIC_CHANNEL, "Nobody is watching this video... Let's reset it.");
			llResetScript();
		}
		if (llGetInventoryNumber(INVENTORY_NOTECARD) <= 0)
		{
			llOwnerSay("Missing notecard!...");
			llResetScript();
		}
		gName = llGetInventoryName(INVENTORY_NOTECARD, 0);
		gLine = 0;
		gQueryID = llGetNotecardLine(gName, gLine);
	}

	dataserver(key query_id, string data)
	{
		if (query_id == gQueryID)
		{
			if (data != EOF)
			{
				llInstantMessage(avatarKey, data);
				++gLine;
				gQueryID = llGetNotecardLine(gName, gLine);
			} else {
				// EOF reached; cleaning up
				gLine = 0;

				string request = "token=" + token
					+ "&avatarName=" + avatarName
					+ "&avatarKey=" + (string)avatarKey;
				reqDelete = llHTTPRequest(streamerAPI + "/delete", [
						HTTP_METHOD, "POST",
						HTTP_MIMETYPE, "application/x-www-form-urlencoded",
						HTTP_ACCEPT, "text/plain", // avoid HTML, since we might not have enough memory to display that
						HTTP_VERBOSE_THROTTLE, FALSE
					],
					request);
				llSetTimerEvent(30.0);	// should be enough to wait
			}
		}
	}

	changed(integer c)
	{
		if (c & CHANGED_INVENTORY)
		{
			llResetScript();
		}
	}

	timer()
	{
		// timeout trying to delete token on the server, so we just clean up.
		llSetTimerEvent(0.0);
		token = ""; avatarName = ""; avatarKey = NULL_KEY;
		state default;
	}

	http_response(key request_id, integer status, list metadata, string body)
	{
		if (request_id == NULL_KEY)
		{
			llRegionSay(BT_DEBUG_CHANNEL, "Weird, a null HTTP request received just now...");
		}
		else if (request_id == reqDelete)
		{
			if (status == 200)
			{
				// Token was deleted, wrap it up!
				llRegionSay(BT_DEBUG_CHANNEL, body);
			}
			else
			{
				llRegionSay(BT_DEBUG_CHANNEL, "Error while deleting token '" + token + "', error - " + (string)status + ": " + body);
			}
		}
		// we wrap it up anyway
		token = ""; avatarName = ""; avatarKey = NULL_KEY;
		state default;
	}
}