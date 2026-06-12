import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 50 },  
    { duration: '20s', target: 150 }, 
    { duration: '10s', target: 0 },   
  ],
};

const MY_JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODEyODE2MDgsInRva2VuVmVyc2lvbiI6MSwidXNlcklEIjoyLCJ1c2VyTmFtZSI6Ikl0dWpoIn0.UcU2y0BFVAbHele0nEDLAWmIR-WCENDmPEb-FETh9Cc";

export default function () {
    // Внутри функции default function() в test.js:
    const randomBool = Math.random() < 0.5;
    const randomId = Math.floor(Math.random() * 1000); 

    const url = `http://localhost:1337/songs/?start=0&count=20&sorted=${randomBool}&search=${randomId}`;

  const params = {
    headers: {
      'Authorization': `Bearer ${MY_JWT_TOKEN}`,
      'Content-Type': 'application/json',
    },
  };

  http.get(url, params);
  
  sleep(0.1); 
}
