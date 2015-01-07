package account

import (
	//"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateName(t *testing.T) {

	//Create account
	Create("TestPerson", "test@person.com", "Password!234123", "en-US", true)

	Convey("Validate name", t, func() {
		assert.Nil(t, ValidateName("thompson", "en-US"), "Name is valid.")
		assert.Error(t, ValidateName("Test", "en-US"), "Name is too short.")
		assert.Error(t, ValidateName("testperson", "en-US"), "Name is already exists.")
	})

	//Delete account
	Delete("test@person.com", "en-US")
}

func TestValidateEmail(t *testing.T) {

	//Create account
	Create("TestEmail", "test@email.com", "Password!234123", "en-US", true)

	Convey("Validate email", t, func() {
		assert.Nil(t, ValidateEmail("thompson@reuters.com", "en-US"))
		assert.Error(t, ValidateEmail("test@emai", "en-US"), "Invalid email address.")
		assert.Error(t, ValidateEmail("test@email.com", "en-US"), "Email already in used.")
	})

	//Delete account
	Delete("test@email.com", "en-US")
}

func TestValidatePassword(t *testing.T) {

	Convey("Validate password", t, func() {
		assert.Nil(t, ValidatePassword("Password!234", "en-US"))
		assert.Error(t, ValidatePassword("Passwo", "en-US"), "Password is not strong enough.")
	})
}

func TestCreatAccount(t *testing.T) {

	Convey("Create account", t, func() {
		assert.Nil(t, Create("Thompson", "thompson@reueters.com", "Password!1232", "en-US", true))
	})

}

func TestDeleteAccount(t *testing.T) {

	Convey("Delete account", t, func() {
		assert.Nil(t, Delete("thompson@reueters.com", "en-US"))
		assert.Error(t, Delete("not@available.com", "en-US"), "Account is not available.")
	})
}

func TestReadAccount(t *testing.T) {

	//Create account
	Create("TestRead", "testread@email.com", "Password!234123", "en-US", true)

	Convey("Read account", t, func() {
		user, _ := Read("testread@email.com", "en-US")
		assert.Equal(t, user.Email, "testread@email.com")
		assert.Equal(t, user.Name, "testread")
	})

	//Delete account
	Delete("testread@email.com", "en-US")
}

func TestUpdateAccount(t *testing.T) {

	//Create account
	Create("testupdate", "testupdate@email.com", "Password!234123", "en-US", true)

	Convey("Update account", t, func() {

		assert.Nil(t, Update("testupdate@email.com", "Vacation", "testupdate2@email.com", "en-US"), "Update account.")

		err := Update("testupdate2@email.com", "Inva", "testupdaer.", "en-US")
		assert.Equal(t, err[0].Error(), "Email address is not valid.")
		assert.Equal(t, err[1].Error(), "Name is either too short or too long.")

	})

	//Delete account
	Delete("testupdate2@email.com", "en-US")
}

// func TestActivateAccount(t *testing.T) {

// 	//Create account
// 	//Already activated account, for testing, if would be return error
// 	//Create("activated", "activated@email.com", "Password!234123", "en-US", true)
// 	Create("toactivated", "toactivated@email.com", "Password!234123", "en-US", true)

// 	Convey("Update account", t, func() {
// 		assert.Error(t, Activate("activated@email.com", "en-US"), "Account has been already activated.")
// 		assert.Nil(t, Activate("toactivated@email.com", "en-US"), "Activate account.")
// 	})

// 	//Delete account
// 	Delete("toactivated@email.com", "en-US")
// }

func TestCloseAccount(t *testing.T) {

	//Create account
	Create("toclose", "toclose@email.com", "Password!234123", "en-US", true)

	Convey("Update account", t, func() {
		assert.Nil(t, Close("toclose@email.com", "Posted bad picture.", "en-US"), "Close account.")
	})

	//Delete account
	Delete("toclose@email.com", "en-US")
}

func TestSignIn(t *testing.T) {

	//Create account
	Create("signinat", "signinat@email.com", "Password!234123", "en-US", true)

	Convey("Record signin date", t, func() {
		_, err := SignIn("signinat@email.com", "Password!234123", "en-US")
		assert.NoError(t, err, "Sign in successfully.")
		_, err = SignIn("signinat@email.com", "Password!34", "en-US")
		assert.Error(t, err, "Sign in fail.")
	})

	//Delete account
	Delete("signinat@email.com", "en-US")

}

func TestFilterDirtyName(t *testing.T) {

	Convey("Filter dirty name.", t, func() {
		assert.Nil(t, filterDirtyName("zerocoding", "en-US"), "Name is allowed.")
		assert.Error(t, filterDirtyName("motherfucker", "en-US"), "Name is prohibited.")
		assert.Error(t, filterDirtyName("iamkiller", "en-US"), "Name is prohibited.")
		assert.Error(t, filterDirtyName("iammilfhunter", "en-US"), "Name is prohibited.")
	})

}

func TestChangePassword(t *testing.T) {

	//Create account
	Create("changepassword", "changepassword@email.com", "Password!234123", "en-US", true)

	Convey("Change password.", t, func() {
		assert.Error(t, ChangePassword("changepassword@email.com", "Password!2", "Teste!123", "en-US"), "Wrong old password.")
		assert.Error(t, ChangePassword("changepassword@email.com", "Password!234123", "Teste", "en-US"), "New password does not match security requirements.")
		assert.Nil(t, ChangePassword("changepassword@email.com", "Password!234123", "Test!234", "en-US"), "Change password successfully.")
	})

	//Delete account
	Delete("changepassword@email.com", "en-US")

}

func TestResetPassword(t *testing.T) {

	//Create account
	Create("resetpassword", "resetpassword@email.com", "Password!2323", "en-US", true)

	Convey("Change password.", t, func() {
		assert.Error(t, ResetPassword("resetpassword@email.com", "Tes", "en-US"), "Password is not string enough.")
		assert.Nil(t, ResetPassword("resetpassword@email.com", "Rest!2344", "en-US"), "Password is not string enough.")
	})

	//Delete account
	Delete("resetpassword@email.com", "en-US")

}

func TestUpdateEmail(t *testing.T) {

	//Create account
	Create("updateemail", "update@email.com", "Password!2323", "en-US", true)

	Convey("Update email.", t, func() {
		UpdateEmail("update@email.com", "update2@email.com", "en-US")
		user, err := Read("update2@email.com", "en-US")
		assert.NoError(t, err, "Should not contain any errors.")
		assert.Equal(t, "update2@email.com", user.Email, "Should update the email address.")
	})

	//Delete account
	Delete("update2@email.com", "en-US")

}

func TestUpdateName(t *testing.T) {

	//Create account
	Create("updatename", "update@name.com", "Password!2323", "en-US", true)

	Convey("Update name.", t, func() {
		UpdateName("update@name.com", "updatename2", "en-US")
		user, err := Read("update@name.com", "en-US")
		assert.NoError(t, err, "Should not contain any errors.")
		assert.Equal(t, "updatename2", user.Name, "Should update the name.")
	})

	//Delete account
	Delete("update@name.com", "en-US")

}
