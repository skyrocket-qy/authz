import http from 'k6/http';


export const options = {
    vus: 1,
    setupTimeout: '6000s',
    iterations: 1,
}

let enforcements = [];

export function setup() {
    let roles     = 10000
    let resources = 1000
    let users     = 100000

    for (let i = 0; i < users; i++){
        let roleId = i % roles
        assignRole(i,roleId)
    }

    for (let i = 0; i < roles; i++){
        let resId = i % resources
        grantPerm(i, resId, "read")
    }


    for (let i = 0; i < 17; i++) {
        const userNum = Math.floor(users / 17) * i;
        const roleNum = userNum % roles;
        let resourceNum = roleNum % resources;

        if (i % 2 === 0) {
            resourceNum = (resourceNum + 1) % resources;
        }

        enforcements.push({
            userId: userNum,
            resourceId: resourceNum,
            perm: "read",
        });
    }

    return { enforcements };
}

export default function (data){
    const idx = __ITER % data.enforcements.length;
    const check = data.enforcements[idx];
    checkPerm(check.userId,check.resourceId,check.perm);
}

// export function teardown(data) {
//   ClearAll();
// }

export function assignRole(userId, roleId) {
  return http.post('http://localhost:8080/authzpb.rbacpb.RbacService/AssignRole', JSON.stringify({ userId, roleId }), {
    headers: { 'Content-Type': 'application/json' },
  });
}

export function grantPerm(roleId, resourceId, perm) {
  return http.post('http://localhost:8080/authzpb.rbacpb.RbacService/GrantPerm', JSON.stringify({ roleId, resourceId, perm }), {
    headers: { 'Content-Type': 'application/json' },
  });
}

export function checkPerm(userId, resourceId, perm) {
    return http.post('http://localhost:8080/authzpb.v1.AuthzService/Check', JSON.stringify({
        sbjId: userId.toString(),
        sbjNs: 'user',
        rel: perm,
        objId: resourceId.toString(),
        objNs: 'object',
    }), { headers: { 'Content-Type': 'application/json' } });
}