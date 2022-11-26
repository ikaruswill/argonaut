use std::fs;
use std::process::Command;
use clap::Parser;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Argo CD Hostname
    #[arg(short = 'a', long)]
    argocd_host: String,
    
    /// Argo CD Application to diff
    #[arg(short = 'n', long)]
    app_name: String,

    /// Commit hash to diff against live
    #[arg(short, long)]
    revision: String,

    /// Gitlab Hostname
    #[arg(short = 'g', long)]
    gitlab_host: String,

    /// Gitlab Project ID
    #[arg(short = 'p', long)]
    gitlab_project_id: String,

    /// Gitlab Merge Request ID
    #[arg(short = 'm', long)]
    gitlab_mr_id: String,
}

fn main() {
    println!("Starting Argonaut");

    let args = Args::parse();

    let argocd_hostname = &args.argocd_host;
    let argocd_token = get_argocd_token();
    let app_name = &args.app_name;
    let revision = &args.revision;

    let diff = get_argocd_diff(argocd_hostname, &argocd_token, app_name, revision);
    println!("{diff}");

    let gitlab_hostname = &args.gitlab_host;
    let gitlab_token = get_gitlab_token();
    let project_id = &args.gitlab_project_id;
    let mr_id = &args.gitlab_mr_id;
    post_gitlab_comment(
        &gitlab_token,
        gitlab_hostname,
        project_id,
        mr_id,
        &diff,
    );
}

fn get_argocd_diff(hostname: &str, token: &str, app_name: &str, revision: &str) -> String {
    let cmd_stdout = Command::new("argocd")
        .arg("app")
        .arg("diff")
        .arg("--server")
        .arg(hostname)
        .arg("--auth-token")
        .arg(token)
        .arg(app_name)
        .arg("--revision")
        .arg(revision)
        .output()
        .expect("Failed")
        .stdout;
    let raw_diff = String::from_utf8_lossy(&cmd_stdout).to_string();
    return raw_diff.trim().to_string();
}

fn get_argocd_token() -> String {
    let token = fs::read_to_string("token-argocd.txt").expect("Unable to read file");
    return token.trim().to_string();
}

fn get_gitlab_token() -> String {
    let token = fs::read_to_string("token-gitlab.txt").expect("Unable to read file");
    return token.trim().to_string();
}

fn post_gitlab_comment(
    token: &str,
    gitlab_server: &str,
    project_id: &str,
    mr_id: &str,
    text: &String,
) {
    let endpoint = format!(
        "https://{gitlab_server}/api/v4/projects/{project_id}/merge_requests/{mr_id}/discussions"
    );
    let body = format!(
        r#"
```diff
{text}
```"#
    );
    let client = reqwest::blocking::Client::new();
    let res = client
        .post(&endpoint)
        .header("PRIVATE-TOKEN", token)
        .query(&[("body", body)])
        .send()
        .expect("Failed during request");
    let res_text = res.text().expect("Failed to get text");
    println!("{res_text}")
}
