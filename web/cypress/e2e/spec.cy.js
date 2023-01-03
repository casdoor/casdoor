describe('登录', () => {
  it('登陆成功', () => {
    cy.visit('http://localhost:7001')
    cy.request({
      method: "POST",
      url: "http://localhost:7001/api/login",
      body:{
        "application": "app-built-in",
        "organization": "built-in",
        "username": "admin",
        "password": "123",
        "autoSignin": true,
        "type": "login",
        "phonePrefix": "86"
      }
    }).then((Response)=>{
      console.log(Response.body)
      expect(Response).property('body').property('status').to.equal('ok')
    })
  });
  it('ui登陆成功', () => {
    cy.visit('http://localhost:7001')
    cy.get('#input').type("admin");
    cy.get('#normal_login_password').type('123')
    cy.get('.ant-btn').click();
    cy.url().should('eq',"http://localhost:7001/")
  });
  it('登陆失败', () => {
    cy.visit('http://localhost:7001')
    cy.request({
      method: "POST",
      url: "http://localhost:7001/api/login",
      body:{
        "application": "app-built-in",
        "organization": "built-in",
        "username": "admin",
        "password": "1234",
        "autoSignin": true,
        "type": "login",
        "phonePrefix": "86"
      }
    }).then((Response)=>{
      console.log(Response.body)
      expect(Response).property('body').property('status').to.equal('error')
    })
  });
  it('ui登陆失败', () => {
    cy.visit('http://localhost:7001')
    cy.get('#input').type("admin");
    cy.get('#normal_login_password').type('1234')
    cy.get('.ant-btn').click();
    cy.url().should('eq',"http://localhost:7001/login")
  });
})