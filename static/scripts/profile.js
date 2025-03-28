import { AuthService } from "./authservice.js";
import { CommentService } from "./commentservice.js";
import { USER_STATE, recyclebinState } from "./data.js";
import { PostManager } from "./postmanager.js";
import { formatTimeAgo } from "./timestamps.js";
import { toast } from "./toast.js";

function toTitleCase(str) {
	return str?.replace(/\b\w/g, (char) => char.toUpperCase());
}

class ProfileDashboard {
	constructor() {
		this.authService = new AuthService();
		this.commentService = new CommentService();
		this.postManager = new PostManager();
		this.userData = null;
	}
}

ProfileDashboard.prototype.init = async function () {
	const userData = await this.authService.userDashboard();

	if (userData.error) {
		alert(userData.message);
		window.location.href = "/auth";
		return;
	}

	if (userData.data) {

		USER_STATE.profilePic = userData.data.user_info?.image;
		USER_STATE.posts = userData.data.posts || [];
		USER_STATE.likedPosts = userData.data.liked_posts || [];
		USER_STATE.userComments = userData.data.comments || [];
		USER_STATE.activities = userData.data.activities || [];
		USER_STATE.likes = userData.data.likes;
		USER_STATE.dislikes = userData.data.dislikes;
		USER_STATE.replies = userData.data.replies;
		USER_STATE.bio = userData.data.user_info?.bio;
		USER_STATE.username = userData.data.user_info?.user_name;

		this.userData = userData.data.user_info;
	}

	this.cacheElements();
	this.setupEventListeners();
	this.updateTheme();
	this.updateStats();
	this.renderActivities();
	this.renderComments();
	this.updateActiveSection();
};

ProfileDashboard.prototype.createCommentHTML = function (comment) {
	return `
        <div class="comment-item" data-comment-id="${comment.comment_id}">
			<div class="comment-user-actions">
				<button class="edit-button" id="editCommentBtn" data-comment-id="${
					comment.comment_id
				}">
              		<i data-lucide="edit"></i>
            	</button>
				<button class="delete-button" id="deleteCommentBtn" data-comment-id="${
					comment.comment_id
				}">
                    <i data-lucide="trash-2"></i>
            	</button>
			</div>
            <div class="comment-content"> 
                <div class="profile-image">
                    <img src="${comment.author_img}" 
                         onerror="this.onerror=null;this.src='/static/profiles/avatar.jpg';"/>
                </div>
                <div>
                    <div class="comment-author">${comment.user_name}</div> 
                    <div class="comment-text">${comment.comment}</div> 
                </div>
				<div>|</div>
				<div class="profile-image">
                    <img src="${comment.post_author_img}" 
                         onerror="this.onerror=null;this.src='/static/profiles/avatar.jpg';"/>
                </div>
				<div>
                    <div class="comment-author">${comment.post_title}</div> 
                    <div class="comment-text"> by ${comment.post_author}</div> 
                </div>
            </div>
            <div class="comment-footer">
                <div class="comment-actions"> 
                    <button class="comment-action-button like-button data-comment-id="${
											comment.comment_id
										}"> 
                        <i data-lucide="thumbs-up"></i> 
                        <span class="likes-count">${
													comment.likes?.length || 0
												}</span> 
                    </button>
					 <button class="comment-action-button dislike-button data-comment-id="${
							comment.comment_id
						}"> 
                        <i data-lucide="thumbs-down"></i> 
                        <span class="likes-count">${
													comment.dislikes?.length || 0
												}</span> 
                    </button>
                </div>
                <div class="comment-meta">
                    <span class="comment-time">${formatTimeAgo(
											comment.created_at
										)}</span> 
                </div>
            </div>
        </div>`;
};

ProfileDashboard.prototype.renderComments = function () {
	this.commentsList = document.getElementById("commentsList");

	let comments = ``;

	USER_STATE.userComments?.forEach((comment) => {
		comments += this.createCommentHTML(comment);
	});

	this.commentsList.innerHTML = comments;
	lucide.createIcons();

	document.querySelectorAll("#editCommentBtn").forEach((button) => {
		button.addEventListener("click", (e) => this.editComment(e));
	});

	document.querySelectorAll("#deleteCommentBtn").forEach((button) => {
		button.addEventListener("click", (e) => this.deleteComment(e));
	});
};

ProfileDashboard.prototype.editComment = async function (e) {
	const button = e.currentTarget.closest("#editCommentBtn");
	if (!button) return;
	const commentId = button.getAttribute("data-comment-id");

	const comment = USER_STATE.userComments?.find(
		(comment) => comment.comment_id === commentId
	);

	const editComment = prompt("Edit this comment?", comment.comment);
	if (!editComment) return;

	comment.comment = editComment;

	const res = await this.commentService.updateComment(comment);
	if (res.error) {
		toast.createToast("error", res.message);
		return;
	}

	if (res.data) {
		toast.createToast("success", res.message || "Comment updated!");
	}
};

ProfileDashboard.prototype.deleteComment = async function (e) {
	const button = e.currentTarget.closest("#deleteCommentBtn");
	if (!button) return;
	const commentId = button.getAttribute("data-comment-id");

	const comment = USER_STATE.userComments?.find(
		(comment) => comment.comment_id === commentId
	);

	const res = await this.commentService.deleteComment(comment);
	if (res.error) {
		toast.createToast("error", res.message);
		return;
	}

	toast.createToast("success", "Comment deleted!");
};

ProfileDashboard.prototype.cacheElements = function () {
	this.elements = {
		userName: document.getElementById("username"),
		profileImage: document.getElementById("profileImage"),
		headerImage: document.getElementById("userProfileImage"),
		imageUpload: document.getElementById("imageUpload"),
		bioText: document.getElementById("bioText"),
		editBioButton: document.getElementById("editBioButton"),
		darkModeToggle: document.getElementById("darkModeToggle"),
		sections: {
			overview: document.getElementById("overviewSection"),
			posts: document.getElementById("postsSection"),
			likedPosts: document.getElementById("likedPostsSection"),
			comments: document.getElementById("commentsSection"),
			settings: document.getElementById("settingsSection"),
		},
		sidebarItems: document.querySelectorAll(".sidebar-item"),
	};
	this.elements.bioText.textContent =
		USER_STATE.bio || "Hey there! I'm on forum.";
	this.elements.profileImage.src =
		USER_STATE.profilePic ||
		"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%239ca3af'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z'/%3E%3C/svg%3E";
	this.elements.userName.textContent = toTitleCase(USER_STATE.username);
};

ProfileDashboard.prototype.setupEventListeners = function () {
	this.elements.darkModeToggle.addEventListener(
		"click",
		this.toggleDarkMode.bind(this)
	);
	this.elements.imageUpload.addEventListener(
		"change",
		this.handleImageUpload.bind(this)
	);
	this.elements.editBioButton.addEventListener(
		"click",
		this.editBio.bind(this)
	);
	this.elements.sidebarItems.forEach((item) => {
		item.addEventListener("click", () => this.switchView(item.dataset.view));

		const data = item.getAttribute("data-view");

		if (data === "likedPosts")
			item.addEventListener("click", this.renderLikedPosts.bind(this));
		if (data === "posts")
			item.addEventListener("click", this.renderMyPosts.bind(this));
	});
};

ProfileDashboard.prototype.switchView = function (view) {
	USER_STATE.currentView = view;
	this.updateActiveSection();
};

ProfileDashboard.prototype.updateActiveSection = function () {
	// Hide all sections
	Object.values(this.elements.sections).forEach((section) =>
		section.classList.add("hidden")
	);

	this.elements.sections[USER_STATE.currentView].classList.remove("hidden");
	this.elements.sidebarItems.forEach((item) => item.classList.remove("active"));

	const activeItem = Array.from(this.elements.sidebarItems).find(
		(item) => item.dataset.view === USER_STATE.currentView
	);
	if (activeItem) activeItem.classList.add("active");
};

ProfileDashboard.prototype.renderMyPosts = function (e) {
	e.preventDefault();

	recyclebinState.RECYCLEBIN = null;
	this.postManager.renderPosts(USER_STATE.posts);
};

ProfileDashboard.prototype.renderLikedPosts = function (e) {
	e.preventDefault();

	recyclebinState.RECYCLEBIN = "items";
	this.postManager.renderPosts(USER_STATE.likedPosts);
};

ProfileDashboard.prototype.toggleDarkMode = function () {
	USER_STATE.darkMode = !USER_STATE.darkMode;
	localStorage.setItem("darkMode", USER_STATE.darkMode);
	this.updateTheme();
};

ProfileDashboard.prototype.updateTheme = function () {
	document.body.setAttribute(
		"data-theme",
		USER_STATE.darkMode ? "dark" : "light"
	);
	this.elements.darkModeToggle.innerHTML = `<i data-lucide="${
		USER_STATE.darkMode ? "sun" : "moon"
	}"></i>`;
	lucide.createIcons();
};

ProfileDashboard.prototype.editBio = async function () {
	const newBio = prompt("Edit your bio:", USER_STATE.bio);
	if (newBio) {
		const formData = {
			user_id: this.userData.user_id,
			bio: newBio,
		};

		const res = await this.authService.editBio(formData);

		if (res.error) {
			toast.createToast("error", res.message);
			return;
		}

		if (res.data) {
			toast.createToast("success", res.message || "Bio updated!");
			USER_STATE.bio = res.data.bio;
			this.elements.bioText.textContent = res.data.bio;
		}
	}
};

ProfileDashboard.prototype.handleImageUpload = async function (e) {
	const file = e.target.files[0];

	if (!file) return;

	// Validate file types and size
	const ALLOWED_TYPES = ["image/jpeg", "image/png", "image/gif"];
	const maxFileSize = 5 * 1024 * 1024; // 5MB

	if (!ALLOWED_TYPES.includes(file.type)) {
		alert("Invalid file type. Please upload a JPEG, PNG, or GIF image.");
		this.elements.imageUpload.value = ""; // Clear the input
		return;
	}

	if (file.size > maxFileSize) {
		alert(
			`Image size is too large.Please upload an image less than ${
				maxFileSize / 1024 / 1024
			} MB.`
		);
		this.elements.imageUpload.value = ""; // Clear the input
		return;
	}

	// Read the file and display it immediately
	const reader = new FileReader();
	reader.onloadend = () => {
		// Update the profile picture immediately
		USER_STATE.profilePic = reader.result;
		this.elements.profileImage.src = reader.result;
	};
	reader.readAsDataURL(file);

	// Upload the file to the server
	const formData = new FormData();
	formData.append("image", file);

	const user = await this.authService.uploadProfilePic(formData);

	if (user.error) {
		toast.createToast("error", user.message);
		return;
	}
	
	if (user.data !== null) {
		USER_STATE.profilePic = user.data.image;
		this.elements.profileImage.src =
			user.data.image || "/static/profiles/avatar.jpg";
		this.elements.headerImage.src = user.data.image;
		toast.createToast("success", "Profile picture updated successfully!");
	}
};

ProfileDashboard.prototype.updateStats = function () {
	document.getElementById("postsCount").textContent =
		USER_STATE.posts?.length || 0;
	document.getElementById("commentsCount").textContent =
		USER_STATE.userComments?.length || 0;
	document.getElementById("likesCount").textContent =
		USER_STATE.likes?.length || 0;
	document.getElementById("dislikeCount").textContent =
		USER_STATE.dislikes?.length || 0;
	document.getElementById("repliesCount").textContent =
		USER_STATE.replies?.length || 0;
};

ProfileDashboard.prototype.renderActivities = function () {
	document.getElementById("activityList").innerHTML = USER_STATE.activities
		?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
		.slice(0, 6)
		.map(
			(activity) => `
		<div class="activity-item">
			<i data-lucide="clock"></i>
			<span>${activity.activity_data}</span> - 
			<span>${formatTimeAgo(activity.created_at)}</span>
		</div>`
		)
		.join(" ");
	lucide.createIcons();
};

const dashboard = new ProfileDashboard();
document.addEventListener("DOMContentLoaded", () => dashboard.init());
