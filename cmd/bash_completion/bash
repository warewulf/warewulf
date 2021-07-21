# bash completion for wwctl                                -*- shell-script -*-

__wwctl_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__wwctl_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__wwctl_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__wwctl_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__wwctl_handle_go_custom_completion()
{
    __wwctl_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly wwctl allows to handle aliases
    args=("${words[@]:1}")
    requestComp="${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __wwctl_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __wwctl_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __wwctl_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __wwctl_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __wwctl_debug "${FUNCNAME[0]}: the completions are: ${out[*]}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __wwctl_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __wwctl_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __wwctl_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out[*]}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __wwctl_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subDir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out[0]}")
        if [ -n "$subdir" ]; then
            __wwctl_debug "Listing directories in $subdir"
            __wwctl_handle_subdirs_in_dir_flag "$subdir"
        else
            __wwctl_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out[*]}" -- "$cur")
    fi
}

__wwctl_handle_reply()
{
    __wwctl_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __wwctl_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __wwctl_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __wwctl_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
		if declare -F __wwctl_custom_func >/dev/null; then
			# try command name qualified custom func
			__wwctl_custom_func
		else
			# otherwise fall back to unqualified for compatibility
			declare -F __custom_func >/dev/null && __custom_func
		fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__wwctl_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__wwctl_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__wwctl_handle_flag()
{
    __wwctl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __wwctl_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __wwctl_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __wwctl_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __wwctl_contains_word "${words[c]}" "${two_word_flags[@]}"; then
			  __wwctl_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__wwctl_handle_noun()
{
    __wwctl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __wwctl_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __wwctl_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__wwctl_handle_command()
{
    __wwctl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_wwctl_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __wwctl_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__wwctl_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __wwctl_handle_reply
        return
    fi
    __wwctl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __wwctl_handle_flag
    elif __wwctl_contains_word "${words[c]}" "${commands[@]}"; then
        __wwctl_handle_command
    elif [[ $c -eq 0 ]]; then
        __wwctl_handle_command
    elif __wwctl_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __wwctl_handle_command
        else
            __wwctl_handle_noun
        fi
    else
        __wwctl_handle_noun
    fi
    __wwctl_handle_word
}

_wwctl_configure_dhcp()
{
    last_command="wwctl_configure_dhcp"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--show")
    flags+=("-s")
    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_configure_hosts()
{
    last_command="wwctl_configure_hosts"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--show")
    flags+=("-s")
    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_configure_nfs()
{
    last_command="wwctl_configure_nfs"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--show")
    flags+=("-s")
    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_configure_ssh()
{
    last_command="wwctl_configure_ssh"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--show")
    flags+=("-s")
    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_configure_tftp()
{
    last_command="wwctl_configure_tftp"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--show")
    flags+=("-s")
    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_configure()
{
    last_command="wwctl_configure"

    command_aliases=()

    commands=()
    commands+=("dhcp")
    commands+=("hosts")
    commands+=("nfs")
    commands+=("ssh")
    commands+=("tftp")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_container_build()
{
    last_command="wwctl_container_build"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--force")
    flags+=("-f")
    flags+=("--setdefault")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_container_delete()
{
    last_command="wwctl_container_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_container_exec()
{
    last_command="wwctl_container_exec"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_container_import()
{
    last_command="wwctl_container_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--build")
    flags+=("-b")
    flags+=("--force")
    flags+=("-f")
    flags+=("--setdefault")
    flags+=("--update")
    flags+=("-u")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_container_list()
{
    last_command="wwctl_container_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_container()
{
    last_command="wwctl_container"

    command_aliases=()

    commands=()
    commands+=("build")
    commands+=("delete")
    commands+=("exec")
    commands+=("import")
    commands+=("list")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_kernel_delete()
{
    last_command="wwctl_kernel_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_kernel_import()
{
    last_command="wwctl_kernel_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--node")
    flags+=("-n")
    flags+=("--root=")
    two_word_flags+=("--root")
    two_word_flags+=("-r")
    flags+=("--setdefault")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_kernel_list()
{
    last_command="wwctl_kernel_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_kernel()
{
    last_command="wwctl_kernel"

    command_aliases=()

    commands=()
    commands+=("delete")
    commands+=("import")
    commands+=("list")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node_add()
{
    last_command="wwctl_node_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cluster=")
    two_word_flags+=("--cluster")
    two_word_flags+=("-c")
    flags+=("--discoverable")
    flags+=("--gateway=")
    two_word_flags+=("--gateway")
    two_word_flags+=("-G")
    flags+=("--hwaddr=")
    two_word_flags+=("--hwaddr")
    two_word_flags+=("-H")
    flags+=("--ipaddr=")
    two_word_flags+=("--ipaddr")
    two_word_flags+=("-I")
    flags+=("--netdev=")
    two_word_flags+=("--netdev")
    two_word_flags+=("-N")
    flags+=("--netmask=")
    two_word_flags+=("--netmask")
    two_word_flags+=("-M")
    flags+=("--type=")
    two_word_flags+=("--type")
    two_word_flags+=("-T")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node_console()
{
    last_command="wwctl_node_console"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node_delete()
{
    last_command="wwctl_node_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force=")
    two_word_flags+=("--force")
    two_word_flags+=("-f")
    flags+=("--yes")
    flags+=("-y")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node_list()
{
    last_command="wwctl_node_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--ipmi")
    flags+=("-i")
    flags+=("--long")
    flags+=("-l")
    flags+=("--net")
    flags+=("-n")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node_sensors()
{
    last_command="wwctl_node_sensors"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--full")
    flags+=("-F")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node_set()
{
    last_command="wwctl_node_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--addprofile=")
    two_word_flags+=("--addprofile")
    flags+=("--all")
    flags+=("-a")
    flags+=("--cluster=")
    two_word_flags+=("--cluster")
    two_word_flags+=("-c")
    flags+=("--comment=")
    two_word_flags+=("--comment")
    flags+=("--container=")
    two_word_flags+=("--container")
    two_word_flags+=("-C")
    flags+=("--delprofile=")
    two_word_flags+=("--delprofile")
    flags+=("--discoverable")
    flags+=("--force")
    flags+=("-f")
    flags+=("--gateway=")
    two_word_flags+=("--gateway")
    two_word_flags+=("-G")
    flags+=("--hwaddr=")
    two_word_flags+=("--hwaddr")
    two_word_flags+=("-H")
    flags+=("--init=")
    two_word_flags+=("--init")
    two_word_flags+=("-i")
    flags+=("--ipaddr=")
    two_word_flags+=("--ipaddr")
    two_word_flags+=("-I")
    flags+=("--ipmi=")
    two_word_flags+=("--ipmi")
    flags+=("--ipmigateway=")
    two_word_flags+=("--ipmigateway")
    flags+=("--ipminetmask=")
    two_word_flags+=("--ipminetmask")
    flags+=("--ipmipass=")
    two_word_flags+=("--ipmipass")
    flags+=("--ipmiuser=")
    two_word_flags+=("--ipmiuser")
    flags+=("--ipxe=")
    two_word_flags+=("--ipxe")
    flags+=("--kernel=")
    two_word_flags+=("--kernel")
    two_word_flags+=("-K")
    flags+=("--kernelargs=")
    two_word_flags+=("--kernelargs")
    two_word_flags+=("-A")
    flags+=("--key=")
    two_word_flags+=("--key")
    two_word_flags+=("-k")
    flags+=("--keydel")
    flags+=("--netdefault")
    flags+=("--netdel")
    flags+=("--netdev=")
    two_word_flags+=("--netdev")
    two_word_flags+=("-N")
    flags+=("--netmask=")
    two_word_flags+=("--netmask")
    two_word_flags+=("-M")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    two_word_flags+=("-P")
    flags+=("--root=")
    two_word_flags+=("--root")
    flags+=("--runtime=")
    two_word_flags+=("--runtime")
    two_word_flags+=("-R")
    flags+=("--system=")
    two_word_flags+=("--system")
    two_word_flags+=("-S")
    flags+=("--type=")
    two_word_flags+=("--type")
    two_word_flags+=("-T")
    flags+=("--undiscoverable")
    flags+=("--value=")
    two_word_flags+=("--value")
    flags+=("--yes")
    flags+=("-y")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_node()
{
    last_command="wwctl_node"

    command_aliases=()

    commands=()
    commands+=("add")
    commands+=("console")
    commands+=("delete")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("del")
        aliashash["del"]="delete"
        command_aliases+=("rm")
        aliashash["rm"]="delete"
    fi
    commands+=("list")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("sensors")
    commands+=("set")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_build()
{
    last_command="wwctl_overlay_build"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_chmod()
{
    last_command="wwctl_overlay_chmod"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--noupdate")
    flags+=("-n")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_create()
{
    last_command="wwctl_overlay_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--noupdate")
    flags+=("-n")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_delete()
{
    last_command="wwctl_overlay_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    flags+=("--noupdate")
    flags+=("-n")
    flags+=("--parents")
    flags+=("-p")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_edit()
{
    last_command="wwctl_overlay_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--files")
    flags+=("-f")
    flags+=("--mode=")
    two_word_flags+=("--mode")
    two_word_flags+=("-m")
    flags+=("--noupdate")
    flags+=("-n")
    flags+=("--parents")
    flags+=("-p")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_import()
{
    last_command="wwctl_overlay_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--mode=")
    two_word_flags+=("--mode")
    two_word_flags+=("-m")
    flags+=("--noupdate")
    flags+=("-n")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_list()
{
    last_command="wwctl_overlay_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--long")
    flags+=("-l")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_mkdir()
{
    last_command="wwctl_overlay_mkdir"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--mode=")
    two_word_flags+=("--mode")
    two_word_flags+=("-m")
    flags+=("--noupdate")
    flags+=("-n")
    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay_show()
{
    last_command="wwctl_overlay_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--system")
    flags+=("-s")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_overlay()
{
    last_command="wwctl_overlay"

    command_aliases=()

    commands=()
    commands+=("build")
    commands+=("chmod")
    commands+=("create")
    commands+=("delete")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("del")
        aliashash["del"]="delete"
        command_aliases+=("rm")
        aliashash["rm"]="delete"
    fi
    commands+=("edit")
    commands+=("import")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cp")
        aliashash["cp"]="import"
    fi
    commands+=("list")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("mkdir")
    commands+=("show")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cat")
        aliashash["cat"]="show"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_power_cycle()
{
    last_command="wwctl_power_cycle"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_power_off()
{
    last_command="wwctl_power_off"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_power_on()
{
    last_command="wwctl_power_on"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_power_status()
{
    last_command="wwctl_power_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_power()
{
    last_command="wwctl_power"

    command_aliases=()

    commands=()
    commands+=("cycle")
    commands+=("off")
    commands+=("on")
    commands+=("status")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_profile_add()
{
    last_command="wwctl_profile_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_profile_delete()
{
    last_command="wwctl_profile_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--yes")
    flags+=("-y")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_profile_list()
{
    last_command="wwctl_profile_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_profile_set()
{
    last_command="wwctl_profile_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    flags+=("--cluster=")
    two_word_flags+=("--cluster")
    two_word_flags+=("-c")
    flags+=("--comment=")
    two_word_flags+=("--comment")
    flags+=("--container=")
    two_word_flags+=("--container")
    two_word_flags+=("-C")
    flags+=("--force")
    flags+=("-f")
    flags+=("--gateway=")
    two_word_flags+=("--gateway")
    two_word_flags+=("-G")
    flags+=("--hwaddr=")
    two_word_flags+=("--hwaddr")
    two_word_flags+=("-H")
    flags+=("--init=")
    two_word_flags+=("--init")
    two_word_flags+=("-i")
    flags+=("--ipaddr=")
    two_word_flags+=("--ipaddr")
    two_word_flags+=("-I")
    flags+=("--ipmigateway=")
    two_word_flags+=("--ipmigateway")
    flags+=("--ipminetmask=")
    two_word_flags+=("--ipminetmask")
    flags+=("--ipmipass=")
    two_word_flags+=("--ipmipass")
    flags+=("--ipmiuser=")
    two_word_flags+=("--ipmiuser")
    flags+=("--ipxe=")
    two_word_flags+=("--ipxe")
    two_word_flags+=("-P")
    flags+=("--kernel=")
    two_word_flags+=("--kernel")
    two_word_flags+=("-K")
    flags+=("--kernelargs=")
    two_word_flags+=("--kernelargs")
    two_word_flags+=("-A")
    flags+=("--key=")
    two_word_flags+=("--key")
    two_word_flags+=("-k")
    flags+=("--keydel")
    flags+=("--netdefault")
    flags+=("--netdel")
    flags+=("--netdev=")
    two_word_flags+=("--netdev")
    two_word_flags+=("-N")
    flags+=("--netmask=")
    two_word_flags+=("--netmask")
    two_word_flags+=("-M")
    flags+=("--root=")
    two_word_flags+=("--root")
    flags+=("--runtime=")
    two_word_flags+=("--runtime")
    two_word_flags+=("-R")
    flags+=("--system=")
    two_word_flags+=("--system")
    two_word_flags+=("-S")
    flags+=("--type=")
    two_word_flags+=("--type")
    two_word_flags+=("-T")
    flags+=("--value=")
    two_word_flags+=("--value")
    flags+=("--yes")
    flags+=("-y")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_profile()
{
    last_command="wwctl_profile"

    command_aliases=()

    commands=()
    commands+=("add")
    commands+=("delete")
    commands+=("list")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("set")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_ready()
{
    last_command="wwctl_ready"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_server_reload()
{
    last_command="wwctl_server_reload"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_server_restart()
{
    last_command="wwctl_server_restart"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_server_start()
{
    last_command="wwctl_server_start"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--foreground")
    flags+=("-f")
    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_server_status()
{
    last_command="wwctl_server_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_server_stop()
{
    last_command="wwctl_server_stop"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_server()
{
    last_command="wwctl_server"

    command_aliases=()

    commands=()
    commands+=("reload")
    commands+=("restart")
    commands+=("start")
    commands+=("status")
    commands+=("stop")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_wwctl_root_command()
{
    last_command="wwctl"

    command_aliases=()

    commands=()
    commands+=("configure")
    commands+=("container")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("vnfs")
        aliashash["vnfs"]="container"
    fi
    commands+=("kernel")
    commands+=("node")
    commands+=("overlay")
    commands+=("power")
    commands+=("profile")
    commands+=("ready")
    commands+=("server")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("-d")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_wwctl()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __wwctl_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("wwctl")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function
    local last_command
    local nouns=()

    __wwctl_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_wwctl wwctl
else
    complete -o default -o nospace -F __start_wwctl wwctl
fi

# ex: ts=4 sw=4 et filetype=sh
