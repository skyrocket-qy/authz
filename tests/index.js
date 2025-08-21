import http from 'k6/http';


// const baseUrl = 'http://localhost:8080'
const baseUrl = 'http://localhost:8080'

export const options = {
    vus: 1,
    setupTimeout: '6000s',
    // iterations: 1000,
    duration: '10s',
}

let enforcements = [];

export function setup() {
    let roles     = 10000
    let resources = 1000
    let users     = 100000

    // for (let i = 0; i < users; i++){
    //     let roleId = i % roles
    //     assignRole(i,roleId)
    // }

    // for (let i = 0; i < roles; i++){
    //     let resId = i % resources
    //     grantPerm(i, resId, "read")
    // }

    // for (let i = 0; i < 17; i++) {
    //     const userNum = Math.floor(users / 17) * i;
    //     const roleNum = userNum % roles;
    //     let resourceNum = roleNum % resources;

    //     if (i % 2 === 0) {
    //         resourceNum = (resourceNum + 1) % resources;
    //     }

    //     enforcements.push({
    //         userId: userNum,
    //         resourceId: resourceNum,
    //         perm: "read",
    //     });
    // }

    return { enforcements };
}

export default function(data){
    const userId = Math.floor(Math.random() * 100000);
    const resourceId = Math.floor(Math.random() * 1000);

    doRandomOp(userId, resourceId);
}

export function doRandomOp(userId, resourceId){
    const rand = Math.random() * 100;
    if(rand < 100){                  // 0–93 = 94% read
        checkPerm(userId, resourceId, 'read');
    } else if(rand < 99){           // 94–96 = 3% create
        createTuple({
            SbjNs: 'user',
            SbjId: userId.toString(),
            rel: 'member',
            objId: Math.floor(Math.random() * 100000).toString(),
            objNs: 'role',
        });
    } else {                        // 97–99 = 3% delete
        const randId = Math.floor(Math.random() * 110000);
        deleteTuples([randId]); // or some ID
    }
}

// export function teardown(data) {
//   ClearAll();
// }

export function assignRole(userId, roleId) {
  return http.post(`${baseUrl}/authzpb.rbacpb.RbacService/AssignRole`, 
    JSON.stringify({ userId, roleId }), {
    headers: { 'Content-Type': 'application/json' },
  });
}

export function grantPerm(roleId, resourceId, perm) {
  return http.post(`${baseUrl}/authzpb.rbacpb.RbacService/GrantPerm`, 
    JSON.stringify({ roleId, resourceId, perm }), {
    headers: { 'Content-Type': 'application/json' },
  });
}

export function checkPerm(userId, resourceId, perm) {
    return http.post(`${baseUrl}/authzpb.v1.AuthzService/Check`, JSON.stringify({
        sbjId: userId.toString(),
        sbjNs: 'user',
        rel: perm,
        objId: resourceId.toString(),
        objNs: 'object',
    }), { headers: { 'Content-Type': 'application/json' } });
}

export function createTuple(tuple){
    return http.post(
        `${baseUrl}/authzpb.v1.AuthzService/CreateTuple`, 
        JSON.stringify(tuple), 
        {headers: { 'Content-Type': 'application/json' }}
    );
}

export function deleteTuples(id){
    return http.post(
        `${baseUrl}/authzpb.v1.AuthzService/DeleteTuples`, 
        JSON.stringify({
            "delete_tuple_ids": {
                "ids": [id]
            }
        }), 
        {headers: { 'Content-Type': 'application/json' }}
    );
}