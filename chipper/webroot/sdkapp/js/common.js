// The purpose of this file is to provide common functions that are used by multiple pages (home, bot settings, etc).


async function getSDKInfo() {
  try {
    var response = await fetch("/api-sdk/get_sdk_info");
    console.log(response)
    if (!response.ok) {
      return undefined;
    }
    var data = await response.json();
    return data;
  } catch (error) {
    console.error('Unable to get SDK info:', error);
    throw error;
  }
}

