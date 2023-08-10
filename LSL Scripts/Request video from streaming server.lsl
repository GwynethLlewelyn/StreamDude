/*
 * New tech using "StreamDude" to request a specific file to be streamed.
 */

string streamerURL  = "https://streaming.betatechnologies.info/StreamDude/api";
string lalMasterKey = "WH13TR5QRC4TFH06";

/*
 *  StreamDude works as a two-step process: first, we send the object ID
 */
key reqAuth, reqPlay;

default
{
	state_entry()
	{
		llSetText(llGetObjectDesc(), <0.8,0.6,0.0>, 1.0);
		llSetTouchText("▶︎ Video");
		llSetClickAction(CLICK_ACTION_PLAY);
	}

	touch_start(integer total_number)
	{
		llInstantMessage(llDetectedKey(0), llDetectedName(0) + ", please wait a bit until the video starts!");
	}
}
