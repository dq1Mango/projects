//This file's only purpose is to generate the intial list of participants, it should not be ran under normal usage

import {file, write, fetch} from 'bun';
console.log("Hello via Bun!");

const accountName = "LeboSporks2025"
const bearer = await file("./bearer.txt").text();
/*
try {
  const response = await fetch(`https://api.twitter.com/2/users/by?usernames=${accountName}`, {
    methhod: "GET",
    headers: {"Authorization": `Bearer ${bearer}`}});
  console.log(response)
  const data = await response.json();
  console.log(data);
  
  const id = data["data"][0]["id"]
  write("./id.txt", id) 
   
} catch(error) {

  console.error('Error fetching data:', error);
}
*/
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

const id = await file("./id.txt").text();
var data = [];
var pagination = "";
while (true) {
  try {
    const response = await fetch(`https://api.twitter.com/2/users/${id}/tweets?${pagination}start_time=2024-08-27T00:00:00.000Z`, {
      method: "GET",
      headers: {"Authorization": `Bearer ${bearer}`}
  
    });
    console.log(response);
    const tempData = await response.json();
    data.push(tempData["data"]);
    console.log(tempData);
    pagination = "pagination_token=" + tempData["meta"]["next_token"] + "&";
    console.log(tempData["meta"]["result_count"])
    if (tempData["meta"]["result_count"] < 10 || tempData["meta"]["result_count"] == undefined) {
      break;
    }

    await sleep(15 * 61 * 1000)
    //for development purposes
  } catch(error){
    console.error("Error fetching data:", error);
    break;
  };
};
write("./everything.txt", JSON.stringify(data));
//let players = [];
